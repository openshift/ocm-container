package ocmcontainer

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/openshift/ocm-container/pkg/backplane"
	"github.com/openshift/ocm-container/pkg/deprecation"
	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/openshift/ocm-container/pkg/featureSet/aws"
	"github.com/openshift/ocm-container/pkg/featureSet/certificateAuthorities"
	"github.com/openshift/ocm-container/pkg/featureSet/gcloud"
	"github.com/openshift/ocm-container/pkg/featureSet/jira"
	"github.com/openshift/ocm-container/pkg/featureSet/opsutils"
	"github.com/openshift/ocm-container/pkg/featureSet/osdctl"
	"github.com/openshift/ocm-container/pkg/featureSet/pagerduty"
	"github.com/openshift/ocm-container/pkg/featureSet/persistentHistories"
	"github.com/openshift/ocm-container/pkg/featureSet/persistentImages"
	personalize "github.com/openshift/ocm-container/pkg/featureSet/personalization"
	"github.com/openshift/ocm-container/pkg/featureSet/scratch"
	"github.com/openshift/ocm-container/pkg/ocm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Error string

func (e Error) Error() string { return string(e) }

const (
	defaultConsolePort = 9999

	// TODO: make this template accept a param for the console port
	consolePortLookupTemplate     = `{{(index (index .NetworkSettings.Ports "9999/tcp") 0).HostPort}}`
	containerStateRunningTemplate = `{{.State.Running}}`

	errHomeEnvUnset         = Error("environment variable $HOME is not set")
	errClusterAndDashArgs   = Error("specifying a cluster with --cluster-id and using a `-` in the first argument are mutually exclusive")
	errContainerNotRunning  = Error("container is not running")
	errInspectQueryEmpty    = Error("inspect requires Go template-formatted query")
	errNoResponseFromEngine = Error("the container engine did not return a response")
	errInvalidClusterId     = Error("invalid cluster ID provided")
)

const publishAllWarningMsg = `
Publishing all ports allows any machine with network access to your computer to view potentially 
sensitive data. This is not recommended, especially on untrusted or unknown networks. 
Use this option only with extreme caution.`

type ocmContainer struct {
	engine                       *engine.Engine
	container                    *engine.Container
	cluster                      *cluster
	dryRun                       bool
	BlockingPostStartExecCmds    [][]string
	NonBlockingPostStartExecCmds [][]string
}

func New(cmd *cobra.Command, args []string) (*ocmContainer, error) {
	var err error
	var dryRun bool = viper.GetBool("dry-run")

	o := &ocmContainer{
		dryRun: dryRun,
	}

	o.engine, err = engine.New(viper.GetString("engine"), viper.GetString("pull"), dryRun)
	if err != nil {
		return o, err
	}

	c := engine.ContainerRef{}
	c.Envs = make(map[string]string)
	c.Volumes = []engine.VolumeMount{}

	// Hard-coded values
	c.Privileged = true
	c.RemoveAfterExit = true

	// image, tag, launchOpts, console, personalization
	c, err = parseFlags(c)
	if err != nil {
		return o, err
	}

	log.Debug(fmt.Sprintf("container ref: %+v\n", c))

	home := os.Getenv("HOME")
	if home == "" {
		return o, errHomeEnvUnset
	}

	backplaneConfig, err := backplane.New(home)
	if err != nil {
		return o, err
	}

	// Copy the backplane config into the container Envs
	maps.Copy(c.Envs, backplaneConfig.Env)
	c.Volumes = append(c.Volumes, backplaneConfig.Mounts...)

	ocmConfig, err := ocm.New(viper.GetString("ocm-url"))
	if err != nil {
		return o, err
	}

	maps.Copy(c.Envs, ocmConfig.Env)
	maps.Copy(c.Envs, ocmContainerEnvs())

	// OCM SDK Client
	connection, err := ocmConfig.Config.Connection()
	if err != nil {
		return o, err
	}
	defer connection.Close()

	// Future-proofing this: if -C/--cluster-id is provided for a cluster ID instead of a positional argument,
	// then parseArgs should just treat all positional arguments as the command to run in the container
	// Note: THERE MIGHT NOT BE A CLUSTER ID PROVIDED, and that's okay
	key, command, err := parseArgs(args, viper.GetString("cluster-id"))
	if err != nil {
		return o, err
	}

	if key != "" {
		if !isValidClusterKey(key) {
			return o, errInvalidClusterId
		}

		ocmV1Cluster, err := getCluster(connection, key)
		if err != nil {
			return o, err
		}

		log.Debug("retrieved cluster info from OCM")

		o.cluster = &cluster{
			id:         ocmV1Cluster.ID(),
			uuid:       ocmV1Cluster.ExternalID(),
			name:       ocmV1Cluster.Name(),
			baseDomain: ocmV1Cluster.DNS().BaseDomain(),
		}

		// Get the cluster's HCP or Hive info
		var hcp = &hcp{}
		var hive = &hive{}
		if ocmV1Cluster.Hypershift().Enabled() {
			hcp, err = GetHcp(connection, ocmV1Cluster)
			if err != nil {
				return o, err
			}
			log.Debug("retrieved cluster HCP info from OCM")
		} else {
			hive, err = GetHive(connection, ocmV1Cluster)
			if err != nil {
				return o, err
			}
			log.Debug("retrieved cluster Hive info from OCM")
		}

		o.cluster.hcp = hcp
		o.cluster.hive = hive
		o.cluster.env = populateClusterEnv(ocmV1Cluster, hcp, hive)

		maps.Copy(c.Envs, o.cluster.env)

	}

	if c.Entrypoint != "" {
		// Entrypoint is set above during parseFlags(), but helpful to print here with verbosity
		log.Printf("setting container entrypoint: %s\n", c.Entrypoint)
	}

	if command != "" {
		log.Printf("setting container command: %s\n", command)
		c.Command = command
	}

	// OCM-Container optional features follow:

	// AWS Credentials
	if featureEnabled("aws") {
		awsConfig, err := aws.New(home)
		if err != nil {
			return o, err
		}
		c.Volumes = append(c.Volumes, awsConfig.Mounts...)
	}

	// Optional Certificate Authority Trust mount
	if featureEnabled("certificate-authorities") && viper.IsSet("ca_source_anchors") {
		caConfig, err := certificateAuthorities.New(viper.GetString("ca_source_anchors"))
		if err != nil {
			return o, err
		}
		c.Volumes = append(c.Volumes, caConfig.Mounts...)
	}

	// disable-console-port is deprecated so this we're also checking the new --no-console-port flag
	// This can be simplified when disable-console-port is deprecated and removed
	if featureEnabled("console-port") && !viper.GetBool("disable-console-port") {
		if c.LocalPorts == nil {
			c.LocalPorts = map[string]int{}
		}
		c.LocalPorts["console"] = defaultConsolePort
	}

	// GCloud configuration
	if featureEnabled("gcp") {
		gcloudConfig, err := gcloud.New(home)
		if err != nil {
			return o, err
		}
		c.Volumes = append(c.Volumes, gcloudConfig.Mounts...)
	}

	if featureEnabled("jira") {
		// Jira configuration
		jiraDirRWMount := viper.GetBool("jira_dir_rw")
		jiraConfig, err := jira.New(home, jiraDirRWMount)
		if err != nil {
			return o, err
		}
		maps.Copy(c.Envs, jiraConfig.Env)
		c.Volumes = append(c.Volumes, jiraConfig.Mounts...)
	}

	if featureEnabled("ops-utils") && viper.IsSet("ops_utils_dir") {
		// SRE Ops Bin dir
		opsDir := viper.GetString("ops_utils_dir")
		opsDirRWMount := viper.GetBool("ops_utils_dir_rw")
		if opsDir != "" {
			opsUtilsConfig, err := opsutils.New(opsDir, opsDirRWMount)
			if err != nil {
				return o, err
			}
			c.Volumes = append(c.Volumes, opsUtilsConfig.Mounts...)
		}
	}

	// OSDCTL configuration
	if featureEnabled("osdctl") {
		osdctlConfig, err := osdctl.New(home)
		if err != nil {
			return o, err
		}
		c.Volumes = append(c.Volumes, osdctlConfig.Mounts...)
	}

	// PagerDuty configuration
	if featureEnabled("pagerduty") {
		pagerDutyDirRWMount := viper.GetBool("pagerduty_dir_rw")
		pagerDutyConfig, err := pagerduty.New(home, pagerDutyDirRWMount)
		if err != nil {
			return o, err
		}
		c.Volumes = append(c.Volumes, pagerDutyConfig.Mounts...)
	}

	// Persistent Histories
	if featureEnabled("persistent-histories") && viper.GetBool("enable_persistent_histories") {
		if persistentHistories.DeprecatedConfig() {
			persistentHistoriesConfig, err := persistentHistories.New(home, o.cluster.id)
			if err != nil {
				return o, err
			}
			for k, v := range persistentHistoriesConfig.Env {
				c.Envs[k] = v
			}
			c.Volumes = append(c.Volumes, persistentHistoriesConfig.Mounts...)
		}
	}

	// Persistent container images
	if featureEnabled("persistent-images") {
		persistentImagesConfig, err := persistentImages.New(home)
		if err != nil {
			return o, err
		}
		c.Volumes = append(c.Volumes, persistentImagesConfig.Mounts...)
	}

	// Personalization
	if featureEnabled("personalizations") && viper.GetBool("enable_personalization_mount") {
		personalizationDirOrFile := viper.GetString("personalization_file")
		personalizationRWMount := viper.GetBool("personalization_dir_rw")

		if personalizationDirOrFile != "" {
			personalizationConfig, err := personalize.New(personalizationDirOrFile, personalizationRWMount)
			if err != nil {
				return o, err
			}
			c.Volumes = append(c.Volumes, personalizationConfig.Mounts...)
		}
	}

	// Scratch Dir mount
	if featureEnabled("scratch-dir") && viper.IsSet("scratch_dir") {
		scratchDir := viper.GetString("scratch_dir")
		scratchDirRWMount := viper.GetBool("scratch_dir_rw")
		if scratchDir != "" {
			scratchConfig, err := scratch.New(scratchDir, scratchDirRWMount)
			if err != nil {
				return o, err
			}
			c.Volumes = append(c.Volumes, scratchConfig.Mounts...)
		}
	}

	// Create the actual container
	err = o.CreateContainer(c)
	if err != nil {
		return o, err
	}

	log.Printf("container created with ID: %v\n", o.container.ID)

	log.Infof(
		"copying ocm config into container: %s - %s\n",
		ocmConfig.Env["OCMC_EXTERNAL_OCM_CONFIG"],
		ocmConfig.Env["OCMC_INTERNAL_OCM_CONFIG"],
	)

	ocmConfigSource := ocmConfig.Env["OCMC_EXTERNAL_OCM_CONFIG"]
	ocmConfigDest := fmt.Sprintf("%s:%s", o.container.ID, ocmConfig.Env["OCMC_INTERNAL_OCM_CONFIG"])

	out, err := o.Copy(ocmConfigSource, ocmConfigDest)
	if err != nil {
		return o, err
	}
	if out != "" {
		log.Printf("OCM config copy output: %s\n", out)
	}

	// Future proof for hive/management/service clusters
	initialLogin := o.cluster.id

	fmt.Printf("logging into cluster: %s\n", initialLogin)
	fmt.Printf("(If you don't see a prompt, try pressing enter)\n")

	backplaneLoginCmd := []string{
		"/bin/bash",
		"-c",
		//fmt.Sprintf("source ~/.bashrc.d/14-kube-ps1.bashrc ; ocm backplane login %s", initialLogin),
		fmt.Sprintf("source ~/.bashrc ; ocm backplane login %s ; cluster_function", initialLogin),
	}
	o.BlockingPostStartExecCmds = append(o.BlockingPostStartExecCmds, backplaneLoginCmd)

	return o, nil
}

func (o *ocmContainer) consolePortEnabled() bool {
	_, ok := o.container.Ref.LocalPorts["console"]
	return ok
}

func (o *ocmContainer) newConsolePortMap() error {
	if !o.consolePortEnabled() {
		return nil
	}

	consolePort, err := o.Inspect(consolePortLookupTemplate)
	if err != nil {
		return err
	}

	portMapCmd := []string{
		"/bin/bash",
		"-c",
		fmt.Sprintf("echo \"%v\" > /tmp/portmap", (consolePort)),
	}

	o.BlockingPostStartExecCmds = append(o.BlockingPostStartExecCmds, portMapCmd)

	return nil
}

// ExecPostRunBlockingCmds starts the blocking exec commands stored in the
// *ocmContainer config
// Blocking commands are those that must succeed to ensure a working ocm-container
func (o *ocmContainer) ExecPostRunBlockingCmds() error {
	var err error
	var running bool

	// Setup the console portmap exec if enabled
	err = o.newConsolePortMap()
	if err != nil {
		return err
	}

	running, err = o.Running()
	if err != nil {
		return err
	}

	if !running {
		err = errContainerNotRunning
		return err
	}

	// Executes while blocking attachment to the container
	wg := sync.WaitGroup{}
	for _, c := range o.BlockingPostStartExecCmds {
		wg.Add(1)

		out, err := o.Exec(c)
		if err != nil {
			wg.Done()
			return err
		}
		if out != "" {
			// This is not log output and should not be suppressed by log levels; do not pass through a logger
			log.Printf("blocking post start exec output: %s\n", out)
		}
		wg.Done()
	}
	wg.Wait()

	return err
}

// ExecPostRunNonBlockingCmds starts the non-blocking exec commands stored
// in the *ocmContainer config
// Non-blocking commands are those that may or may not succeed, but are not
// critical to the operation of the container
func (o *ocmContainer) ExecPostRunNonBlockingCmds() {
	var running bool
	var err error

	// Executes without blocking attachment
	out := make(chan string)

	for _, c := range o.NonBlockingPostStartExecCmds {
		running, err = o.Running()
		if err != nil {
			log.Error(err.Error())
		}
		if !running {
			err = errContainerNotRunning
			log.Error(err.Error())
			break
		}

		go o.BackgroundExecWithChan(c, out)
		log.Printf("background exec output: %v\n", <-out)
		close(out)
	}
}

// parseFlags returns the flags as strings or bool values
func parseFlags(c engine.ContainerRef) (engine.ContainerRef, error) {

	c.Tty = true
	c.Interactive = true

	entrypoint := viper.GetString("entrypoint")
	if entrypoint != "" {
		c.Entrypoint = entrypoint
	}

	// This is a deprecated command - the same can be accomplished with engine-specific
	// entrypoint and positional CMD arguments - but we're keeping it for now to socialize it
	exec := viper.GetString("exec")
	if exec != "" {
		deprecation.Print("--exec", "--entrypoint")
		c.Command = exec
		c.Tty = false
		c.Interactive = false
	}

	// Image options
	registry := viper.GetString("registry")
	repository := viper.GetString("repository")
	image := viper.GetString("image")
	tag := viper.GetString("tag")

	i := engine.ContainerImage{
		Registry:   registry,
		Repository: repository,
		Name:       image,
		Tag:        tag,
	}

	c.Image = i

	// Best-effort passing of launch options
	launchOpts := viper.GetString("launch-opts")
	if launchOpts != "" {
		c.BestEffortArgs = append(
			c.BestEffortArgs,
			func(launchOpts string) []string {
				return strings.Split(launchOpts, " ")
			}(launchOpts)...,
		)
	}
	launchOpsVar := viper.GetString("ocm_container_launch_opts")
	if launchOpsVar != "" || os.Getenv("OCM_CONTAINER_LAUNCH_OPTS") != "" {
		deprecation.Print("OCM_CONTAINER_LAUNCH_OPTS", "launch_opts")
		c.BestEffortArgs = append(
			c.BestEffortArgs,
			func(launchOpts string) []string {
				return strings.Split(launchOpts, " ")
			}(launchOpsVar)...,
		)
	}

	if c.BestEffortArgs != nil {
		log.Warn(
			fmt.Sprintf("Attempting best-effort parsing of 'ocm_container_launch_opts' options: %s\n", launchOpts) +
				"Please use '--verbose' to inspect engine commands if you encounter any issues.",
		)
	}

	if viper.GetBool("publish-all-ports") {
		log.Warn(publishAllWarningMsg)
		c.PublishAll = true
	}

	return c, nil
}

// parseArgs takes a slice of strings and returns the clusterID and the command to execute inside the container
func parseArgs(args []string, cluster string) (string, string, error) {
	// These two are future-proofing for removing the cluster from positional arguments
	if cluster != "" && len(args) == 0 {
		return cluster, "", nil
	}

	if cluster != "" && args[0] != "-" {
		return cluster, strings.Join(args, " "), nil
	}

	// This is invalid usage
	if cluster != "" && args[0] == "-" {
		return "", "", errClusterAndDashArgs
	}

	switch {
	case len(args) == 1:
		deprecation.Print("using cluster ids in a positional argument", "--cluster-id")
		return args[0], "", nil
	case len(args) > 1:
		if args[0] == "-" {
			// Consider this a "no cluster" placeholder, and only return arguments
			args[0] = ""
		}

		s := []string{}

		for _, arg := range args[1:] {
			if arg != "--" {
				s = append(s, arg)
			}
		}

		return args[0], strings.Join(s, " "), nil
	}
	return "", "", nil
}

// This is just a wrapper around Create for readability
func (o *ocmContainer) CreateContainer(c engine.ContainerRef) error {
	return o.Create(c)
}

func (o *ocmContainer) Create(c engine.ContainerRef) error {
	log.Info(fmt.Sprintf("creating container with ref: %+v\n", c))
	container, err := o.engine.Create(c)
	if err != nil {
		return err
	}
	o.container = container
	return nil
}

func (o *ocmContainer) Attach() error {
	return o.engine.Attach(o.container)
}

func (o *ocmContainer) BackgroundAttach(args []string) (string, error) {
	log.Debugf("attaching to container with args: %v", args)
	return o.engine.BackgroundAttach(o.container, args)
}

func (o *ocmContainer) Start(attach bool) error {
	return o.engine.Start(o.container, false)
}

func (o *ocmContainer) StartAndAttach() error {
	return o.engine.Start(o.container, true)
}

func (o *ocmContainer) BackgroundExec(args []string) {
	o.engine.Exec(o.container, args)
}

func (o *ocmContainer) BackgroundExecWithChan(args []string, stdout chan string) {
	out, err := o.engine.Exec(o.container, args)
	if err != nil {
		stdout <- err.Error()
	}
	stdout <- out
}

func (o *ocmContainer) Exec(args []string) (string, error) {
	return o.engine.Exec(o.container, args)
}

// Copy takes a source and destination (optionally with a [container]: prefixed)
// and executes a container engine "cp" command with those as arguments
func (o *ocmContainer) Copy(source, destination string) (string, error) {
	s := filepath.Clean(source)
	d := filepath.Clean(destination)

	args := []string{s, d}

	out, err := o.engine.Copy(args...)

	return out, err
}

func (o *ocmContainer) Inspect(query string) (string, error) {

	if query == "" {
		return "", errInspectQueryEmpty
	}

	out, err := o.engine.Inspect(o.container, query)
	if err != nil {
		return out, err
	}

	// \n has to be cut first, or the quotes will be left in the output
	out = strings.Trim(out, "\n")
	out = strings.Trim(out, "\"")
	out = strings.Trim(out, "'")

	return out, err
}

// Enabled converts user-friendly negative flags (--no-something)
// to programmer-friendly positives.
// Eg: --no-something=true on the CLI becomes enabled(something)=false in the code.
func featureEnabled(flag string) bool {
	var negativeFlag string = lookUpNegativeName(flag)
	var enabled bool = !viper.GetBool(negativeFlag)

	// Print a message if we're going to skip enabling a feature
	if !enabled {
		log.Printf("Found '--no-%s' - skipping feature\n", flag)
	}
	return !viper.GetBool(lookUpNegativeName(flag))
}

// lookUpNegativeName converts a positive feature name to a negative CLI flag name
// so it can be looked up from Viper.
func lookUpNegativeName(flag string) string {
	return "no-" + flag
}

// Running returns a boolean indicating if the container is running in that Point In Time
// Keep in mind the state could change at any time
func (o *ocmContainer) Running() (bool, error) {
	running, err := o.Inspect(containerStateRunningTemplate)
	if err != nil {
		return false, err
	}

	if running == "" {
		return false, errNoResponseFromEngine
	}

	b, err := strconv.ParseBool(running)
	if err != nil {
		return false, err
	}

	return b, nil
}
