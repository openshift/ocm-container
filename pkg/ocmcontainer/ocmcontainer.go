package ocmcontainer

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
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
	personalize "github.com/openshift/ocm-container/pkg/featureSet/personalization"
	"github.com/openshift/ocm-container/pkg/featureSet/scratch"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var errClusterAndUnderscoreArgs = errors.New("specifying a cluster with --cluster-id and using an underscore in the first argument are mutually exclusive")

const (
	consolePortLookupTemplate = `{{(index (index .NetworkSettings.Ports "9999/tcp") 0).HostPort}}`
)

type ocmContainer struct {
	engine                       *engine.Engine
	container                    *engine.Container
	BlockingPostStartExecCmds    [][]string
	NonBlockingPostStartExecCmds [][]string
	State                        string
	dryRun                       bool
	verbose                      bool
}

func New(cmd *cobra.Command, args []string) (*ocmContainer, error) {
	var err error
	var verbose bool = verboseOutput(viper.GetBool("verbose"), viper.GetBool("debug"))
	var dryRun bool = viper.GetBool("dry-run")

	o := &ocmContainer{
		verbose: verbose,
		dryRun:  dryRun,
	}

	o.engine, err = engine.New(viper.GetString("engine"), dryRun, verbose)
	if err != nil {
		return o, err
	}

	c := engine.ContainerRef{}
	// image, tag, launchOpts, console, personalization
	c, err = parseFlags(c)
	if err != nil {
		return o, err
	}
	if verbose {
		fmt.Printf("container ref: %+v\n", c)
	}

	// Set up a map for environment variables
	c.Envs = make(map[string]string)

	// Standard OCM container user environment envs
	// Setting the strings to empty will pass them in
	// in the "-e USER" from the environment format
	// TODO: These should go in the envs.go, and perhaps
	// be a range over the viper.AllKeys() cross-referenced with
	// cmd.ManagedFields (configure?)

	c.Envs["OFFLINE_ACCESS_TOKEN"] = viper.GetString("offline_access_token")

	// standard env vars specified as nil strings will be passed to the engine
	// in as "-e VAR" using the value from os.Environ() to the syscall.Exec() call
	c.Envs["USER"] = ""

	c.Volumes = []engine.VolumeMount{}

	// Future-proofing this: if -c is provided for a cluster ID instead of a positional argument,
	// then parseArgs should just treat all positional arguments as the command to run in the container
	cluster, command, err := parseArgs(args, viper.GetString("cluster-id"))
	if err != nil {
		return o, err
	}

	if cluster != "" {
		if o.verbose {
			fmt.Printf("logging into cluster: %s\n", cluster)
		}
		c.Envs["INITIAL_CLUSTER_LOGIN"] = cluster
	}

	if c.Entrypoint != "" && o.verbose {
		// Entrypoint is set above during parseFlags(), but helpful to print here with verbosity
		fmt.Printf("setting container entrypoint: %s\n", c.Entrypoint)
	}

	if command != "" {
		if o.verbose {
			fmt.Printf("setting container command: %s\n", command)
		}
		c.Command = command
	}

	home := os.Getenv("HOME")
	if home == "" {
		return o, fmt.Errorf("error: HOME environment variable not set")
	}

	backplaneConfig, err := backplane.New(home)
	if err != nil {
		return o, err
	}

	// Copy the backplane config into the container Envs
	maps.Copy(backplaneConfig.Env, c.Envs)
	c.Volumes = append(c.Volumes, backplaneConfig.Mounts...)

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
		c.PublishAll = true
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
		maps.Copy(jiraConfig.Env, c.Envs)
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
		pagerDutyConfig, err := pagerduty.New(home)
		if err != nil {
			return o, err
		}
		c.Volumes = append(c.Volumes, pagerDutyConfig.Mounts...)
	}

	// persistentHistories requires the cluster name, and retrieves it from OCM
	// before entering the container, so the --cluster-id must be provided,
	// enable_persistent_histories must be true, and OCM must be configured
	// for the user (outside the container)
	if featureEnabled("persistent-histories") && viper.GetBool("enable_persistent_histories") {
		if persistentHistories.DeprecatedConfig() && cluster != "" {
			persistentHistoriesConfig, err := persistentHistories.New(home, cluster)
			if err != nil {
				return o, err
			}
			for k, v := range persistentHistoriesConfig.Env {
				c.Envs[k] = v
			}
			c.Volumes = append(c.Volumes, persistentHistoriesConfig.Mounts...)
		}
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

	if o.verbose {
		fmt.Printf("container created with ID: %v\n", o.container.ID)
	}

	return o, nil
}

func (o *ocmContainer) consolePortEnabled() bool {
	return o.container.Ref.PublishAll
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
		fmt.Sprintf("echo \"%v\" > /tmp/portmap", strings.Trim(consolePort, "'")),
	}

	o.BlockingPostStartExecCmds = append(o.BlockingPostStartExecCmds, portMapCmd)

	return nil
}

// ExecPostRunBlockingCmds starts the blocking exec commands stored in the
// *ocmContainer config
// Blocking commands are those that must succeed to ensure a working ocm-container
func (o *ocmContainer) ExecPostRunBlockingCmds() error {
	// Setup the console portmap exec if enabled
	o.newConsolePortMap()

	// Exectues while blocking attachment to the container
	wg := sync.WaitGroup{}
	for _, c := range o.BlockingPostStartExecCmds {
		wg.Add(1)
		out, err := o.Exec(c)
		//out, err := o.Exec(strings.Split(c, " "))
		if err != nil {
			return err
		}
		if out != "" {
			fmt.Println(out)
		}
		wg.Done()
	}
	wg.Wait()
	return nil
}

// ExecPostRunNonBlockingCmds starts the non-blocking exec commands stored
// in the *ocmContainer config
// Non-blocking commands are those that may or may not succeed, but are not
// critical to the operation of the container
func (o *ocmContainer) ExecPostRunNonBlockingCmds() {

	// Executes without blocking attachment
	out := make(chan string)

	for _, c := range o.NonBlockingPostStartExecCmds {
		// go o.BackgroundExec(strings.Split(c, " "))
		go o.BackgroundExecWithChan(c, out)
		fmt.Printf("%v: %v\n", c, <-out)
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
		fmt.Printf("Attempting best-effort parsing of 'ocm_container_launch_opts' options: %s\n", launchOpts)
		fmt.Printf("Please use '--verbose' to inspect engine commands if you encounter any issues.\n")
	}

	return c, nil
}

// parseArgs takes a slice of strings and returns the clusterID and the command to execute inside the container
func parseArgs(args []string, cluster string) (string, string, error) {
	// These two are future-proofing for removing the cluster from positional arguments
	if cluster != "" && len(args) == 0 {
		return cluster, "", nil
	}

	if cluster != "" && args[0] != "_" {
		return cluster, strings.Join(args, " "), nil
	}

	// This is invalid usage
	if cluster != "" && args[0] == "_" {
		return "", "", errClusterAndUnderscoreArgs
	}

	switch {
	case len(args) == 1:
		deprecation.Print("using cluster ids in a positional argument", "--cluster-id")
		return args[0], "", nil
	case len(args) > 1:
		if args[0] == "_" {
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
	if o.verbose {
		fmt.Printf("creating container with ref: %+v\n", c)
	}
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

func (o *ocmContainer) Start(attach bool) error {
	if attach {
		return o.engine.StartAndAttach(o.container)
	}

	return o.engine.Start(o.container)
}

func (o *ocmContainer) StartAndAttach() error {
	return o.Start(true)
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
func (o *ocmContainer) Copy(source, destination string) error {
	s := filepath.Clean(source)
	d := filepath.Clean(destination)

	args := fmt.Sprintf("%s:%s", s, d)

	o.engine.Copy("cp", args)

	return nil
}

func verboseOutput(verbose, debug bool) bool {
	if debug {
		deprecation.Print("--debug", "--verbose")
	}
	return verbose || debug
}

func (o *ocmContainer) Inspect(value string) (string, error) {

	if value == "" {
		value = "{{.Id}}"
	}

	out, err := o.engine.Inspect(o.container, value)
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(out, "\n"), err
}

// Enabled converts user-friendly negative flags (--no-something)
// to programmer-friendly positives.
// Eg: --no-something=true on the CLI becomes enabled(something)=false in the code.
func featureEnabled(flag string) bool {
	var verbose bool
	var enabled bool
	var negativeFlag string

	verbose = verboseOutput(viper.GetBool("verbose"), viper.GetBool("debug"))

	negativeFlag = lookUpNegativeName(flag)
	enabled = !viper.GetBool(negativeFlag)

	// Print a message if we're going to skip enabling a feature
	if verbose && !enabled {
		fmt.Printf("Found '--no-%s' - skipping feature\n", flag)
	}
	return !viper.GetBool(lookUpNegativeName(flag))
}

// lookUpNegativeName converts a positive feature name to a negative CLI flag name
// so it can be looked up from Viper.
func lookUpNegativeName(flag string) string {
	return "no-" + flag
}
