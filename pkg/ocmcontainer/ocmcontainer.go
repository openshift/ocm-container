package ocmcontainer

import (
	"fmt"
	"maps"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/openshift/ocm-container/pkg/deprecation"
	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/openshift/ocm-container/pkg/features"
	"github.com/openshift/ocm-container/pkg/ocm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Error string

func (e Error) Error() string { return string(e) }

const (
	containerStateRunningTemplate = `{{.State.Running}}`

	errHomeEnvUnset         = Error("environment variable $HOME is not set")
	errClusterAndDashArgs   = Error("specifying a cluster with --cluster-id and using a `-` in the first argument are mutually exclusive")
	errContainerNotRunning  = Error("container is not running")
	errInspectQueryEmpty    = Error("inspect requires Go template-formatted query")
	errNoResponseFromEngine = Error("the container engine did not return a response")
)

type Runtime struct {
	engine    *engine.Engine
	container *engine.Container
	dryRun    bool
	command   []string

	PostStartExecHooks           [](func(features.ContainerRuntime) error)
	BlockingPostStartExecCmds    [][]string
	NonBlockingPostStartExecCmds [][]string
	preExecCleanupFuncs          []func()
	postExecCleanupFuncs         []func()
	trapped                      bool
}

func New(cmd *cobra.Command, args []string) (*Runtime, error) {
	var err error
	var dryRun = viper.GetBool("dry-run")

	o := &Runtime{
		dryRun:             dryRun,
		PostStartExecHooks: [](func(features.ContainerRuntime) error){},
	}

	o.engine, err = engine.New(viper.GetString("engine"), viper.GetString("pull"), dryRun)
	if err != nil {
		return o, err
	}

	cluster := viper.GetString("cluster-id")

	c := engine.ContainerRef{
		LocalPorts: map[string]int{},
	}
	// Hard-coded values
	c.Privileged = true
	c.RemoveAfterExit = true

	// image, tag, launchOpts, console, personalization
	c, err = parseFlags(c)
	if err != nil {
		return o, err
	}

	log.Debug(fmt.Sprintf("container ref: %+v\n", c))

	// Set up a map for environment variables
	c.EnvMap = ocmContainerEnvs()

	c.Volumes = []engine.VolumeMount{}
	c.Envs = []engine.EnvVar{}

	// these args are already split and checked by the root command
	if len(args) != 0 {
		o.command = args

		if cluster != "" {
			// if the cluster id is set - we need to use our custom entrypoint to log in first
			o.command = append([]string{"/root/.local/bin/cluster-command-entrypoint"}, args...)
		}
	}

	home := os.Getenv("HOME")
	if home == "" {
		return o, errHomeEnvUnset
	}

	ocmConfig, err := ocm.New()
	if err != nil {
		return o, fmt.Errorf("error creating connection to ocm: %v", err)
	}
	o.RegisterPreExecCleanupFunc(func() { ocm.CloseClient() })

	if cluster != "" {
		conn := ocm.GetClient()
		// check if cluster exists to fail fast
		_, err := ocm.GetCluster(conn, cluster)
		if err != nil {
			return o, fmt.Errorf("%v - using ocm-url %s", err, conn.URL())
		}
		log.Printf("logging into cluster: %s\n", cluster)
		c.EnvMap["INITIAL_CLUSTER_LOGIN"] = cluster
	}

	maps.Copy(c.EnvMap, ocmConfig.Env)

	// OCM-Container optional features follow:

	featureOptions, err := features.Initialize()
	if err != nil {
		log.Errorf("There was an error initializing a feature: %v", err)
		os.Exit(2)
	}

	c.Volumes = append(c.Volumes, featureOptions.Mounts...)
	c.Envs = append(c.Envs, featureOptions.Envs...)
	maps.Copy(c.LocalPorts, featureOptions.PortMap)
	o.PostStartExecHooks = append(o.PostStartExecHooks, featureOptions.PostStartExecHooks...)

	// Parse additional mounts from the config file
	if viper.IsSet("volumeMounts") {
		mounts := []engine.VolumeMount{}
		var vols []any
		err := viper.UnmarshalKey("volumeMounts", &vols)
		if err != nil {
			log.Errorf("unable to parse volumeMounts config %s", err)
			os.Exit(10)
		}

		unsupportedMountTypes := false
		for _, vol := range vols {
			if v, ok := vol.(string); ok {
				log.Debugf("Parsing bind mount '%s' as string", v)
				mount, err := parseMountString(v)
				if err != nil {
					log.Errorf("error parsing configured mount string '%s': %v", v, err)
					os.Exit(10)
				}
				mounts = append(mounts, mount)
				continue
			}
			// Here is where we will process additional mounts as a map, if we decide to go that direction:
			//log.Debugf("Parsing bind mount as map '%+v'", vol)
			log.Errorf("unsupported mount: %+v", vol)
			unsupportedMountTypes = true
			continue
		}
		if unsupportedMountTypes {
			os.Exit(10)
		}
		c.Volumes = append(c.Volumes, mounts...)
	}

	// Parse additional mounts if they're passed through the CLI
	if viper.IsSet("vols") {
		mounts := []engine.VolumeMount{}
		for _, mountString := range viper.GetStringSlice("vols") {
			log.Debugf("parsing mount string '%s'", mountString)
			mount, err := parseMountString(mountString)
			if err != nil {
				log.Errorf("error parsing additional mount string '%s': %v", mountString, err)
				os.Exit(10)
			}
			mounts = append(mounts, mount)
		}
		c.Volumes = append(c.Volumes, mounts...)
	}

	// Parse additional environment variables if they're passed from config
	// we use `env` to stay consistent with the kubernetes yaml for pod envs
	if viper.IsSet("env") {
		log.Debug("Parsing Additional Env Vars from Config")
		envs := []engine.EnvVar{}
		var rawEnvs []map[string]string
		err := viper.UnmarshalKey("env", &rawEnvs)
		if err != nil {
			log.Errorf("error parsing additional environment vars: %v", err)
			os.Exit(10)
		}

		for _, e := range rawEnvs {
			env := engine.EnvVar{
				Key:   e["name"],
				Value: e["value"],
			}
			log.Debugf("parsing env: %+v", env)
			envs = append(envs, env)
		}
		c.Envs = append(c.Envs, envs...)
	}

	// Parse additional environment variables from the command line, if they exist
	if viper.IsSet("environment") {
		log.Debug("Parsing additional env vars from CLI Flags")
		envs := []engine.EnvVar{}
		rawEnvs := viper.GetStringSlice("environment")
		log.Debugf("rawEnvs: %+v", rawEnvs)
		for _, e := range rawEnvs {
			log.Debugf("parsing string: %s", e)
			env, err := engine.EnvVarFromString(e)
			if err != nil {
				log.Errorf("error parsing flag-defined env var: %v", err)
				os.Exit(10)
			}
			log.Debugf("parsed env: %+v", env)
			envs = append(envs, env)
		}
		c.Envs = append(c.Envs, envs...)
	}

	// Create the actual container
	err = o.CreateContainer(c)
	if err != nil {
		return o, err
	}

	log.Printf("container created with ID: %v\n", o.container.ID)

	log.Debugf(
		"copying ocm config into container: %s - %s\n",
		ocmConfig.Env["OCMC_EXTERNAL_OCM_CONFIG"],
		ocmConfig.Env["OCMC_INTERNAL_OCM_CONFIG"],
	)

	ocmConfigSource := ocmConfig.Env["OCMC_EXTERNAL_OCM_CONFIG"]
	ocmConfigDest := fmt.Sprintf("%s:%s", o.container.ID, ocmConfig.Env["OCMC_INTERNAL_OCM_CONFIG"])

	out, err := o.Copy(ocmConfigSource, ocmConfigDest)
	log.Debug(out)
	if err != nil {
		return o, err
	}

	return o, nil
}

func (o *Runtime) RegisterBlockingPostStartCmd(cmd []string) {
	o.BlockingPostStartExecCmds = append(o.BlockingPostStartExecCmds, cmd)
}

func (o *Runtime) ExecPostRunBlockingHooks() error {
	log.Debugf("Running Post Run Blocking Hooks: %v", o.PostStartExecHooks)
	for _, f := range o.PostStartExecHooks {
		err := f(o)
		if err != nil {
			return err
		}
	}
	return nil
}

// ExecPostRunBlockingCmds starts the blocking exec commands stored in the
// *Runtime config
// Blocking commands are those that must succeed to ensure a working ocm-container
func (o *Runtime) ExecPostRunBlockingCmds() error {
	var err error
	var running bool

	// execute the hooks
	err = o.ExecPostRunBlockingHooks()
	if err != nil {
		return err
	}

	// Executes while blocking attachment to the container
	wg := sync.WaitGroup{}
	for _, c := range o.BlockingPostStartExecCmds {
		wg.Add(1)
		running, err = o.Running()
		if err != nil {
			wg.Done()
			break
		}
		if !running {
			err = errContainerNotRunning
			wg.Done()
			break
		}

		out, err := o.Exec(c)
		if err != nil {
			wg.Done()
			return err
		}
		if out != "" {
			log.Println(out)
		}
		wg.Done()
	}
	wg.Wait()

	return err
}

// ExecPostRunNonBlockingCmds starts the non-blocking exec commands stored
// in the *Runtime config
// Non-blocking commands are those that may or may not succeed, but are not
// critical to the operation of the container
func (o *Runtime) ExecPostRunNonBlockingCmds() {
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

		// go o.BackgroundExec(strings.Split(c, " "))
		go o.BackgroundExecWithChan(c, out)
		log.Printf("%v: %v\n", c, <-out)
	}
}

// parseFlags returns the flags as strings or bool values
func parseFlags(c engine.ContainerRef) (engine.ContainerRef, error) {

	c.Tty = true
	c.Interactive = true

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
		log.Warn("Publishing all ports can result in any machine with network access to your computer to have the ability to view potentially sensitive customer data. This is not recommended, especially if you're not sure what else is on the network you're working from. Use this option only with extreme caution.")
		c.PublishAll = true
	}

	return c, nil
}

// This is just a wrapper around Create for readability
func (o *Runtime) CreateContainer(c engine.ContainerRef) error {
	return o.Create(c)
}

func (o *Runtime) Create(c engine.ContainerRef) error {
	log.Info(fmt.Sprintf("creating container with ref: %+v\n", c))
	container, err := o.engine.Create(c)
	if err != nil {
		return err
	}
	o.container = container
	return nil
}

func (o *Runtime) Attach() error {
	return o.engine.Attach(o.container)
}

func (o *Runtime) Run() error {
	o.preExecCleanup()

	if len(o.command) != 0 {
		// Stop the container after we exec, if a command is provided
		o.RegisterPostExecCleanupFunc(func() {
			// stop and rm the container immediately
			o.Stop(0)
		})
		// Trap and run cleanup if we get an interrupt signal
		o.Trap()
		err := o.engine.ExecLive(o.container, o.command)

		log.Debug("Stopping container after exec")
		o.postExecCleanup()

		if o.trapped {
			// this is the default ^C exit status
			os.Exit(130)
		}
		return err
	}
	return o.Attach()
}

func (o *Runtime) Stop(timeout int) error {
	return o.engine.Stop(o.container, timeout)
}

func (o *Runtime) Start(attach bool) error {
	return o.engine.Start(o.container, false)
}

func (o *Runtime) StartAndAttach() error {
	return o.engine.Start(o.container, true)
}

func (o *Runtime) BackgroundExec(args []string) {
	// BackgroundExec is a non-blocking exec command that cannot return any output
	_, _ = o.engine.Exec(o.container, args)
}

func (o *Runtime) BackgroundExecWithChan(args []string, stdout chan string) {
	out, err := o.engine.Exec(o.container, args)
	if err != nil {
		stdout <- err.Error()
	}
	stdout <- out
}

func (o *Runtime) Exec(args []string) (string, error) {
	return o.engine.Exec(o.container, args)
}

// Copy takes a source and destination (optionally with a [container]: prefixed)
// and executes a container engine "cp" command with those as arguments
func (o *Runtime) Copy(source, destination string) (string, error) {
	s := filepath.Clean(source)
	d := filepath.Clean(destination)

	args := []string{s, d}

	out, err := o.engine.Copy(args...)

	return out, err
}

func (o *Runtime) Inspect(query string) (string, error) {

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
	var negativeFlag = lookUpNegativeName(flag)
	var enabled = !viper.GetBool(negativeFlag)

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
func (o *Runtime) Running() (bool, error) {
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

func (o *ocmContainer) RegisterPreExecCleanupFunc(f func()) {
	o.preExecCleanupFuncs = append(o.preExecCleanupFuncs, f)
}

func (o *ocmContainer) RegisterPostExecCleanupFunc(f func()) {
	o.postExecCleanupFuncs = append(o.postExecCleanupFuncs, f)
}

func (o *ocmContainer) Trap() {
	// Trap Command Cancellations
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		log.Error("Interrupt caught. Cleaning up...")
		o.trapped = true
	}()
}

func (o *ocmContainer) preExecCleanup() {
	log.Debug("Running registered pre-exec cleanup functions")
	cleanup(o.preExecCleanupFuncs)
}

func (o *ocmContainer) postExecCleanup() {
	log.Debug("Running registered cleanup functions")
	cleanup(o.postExecCleanupFuncs)
}

func cleanup(cfs []func()) {
	l := len(cfs)
	for i, f := range cfs {
		f()
		num := i + 1
		log.Debugf("%d/%d cleanup funcs run...", num, l)
	}
}

func parseMountString(mount string) (engine.VolumeMount, error) {
	// Check for empty string
	if mount == "" {
		return engine.VolumeMount{}, fmt.Errorf("mount string cannot be empty")
	}

	// Split the mount string by colons
	parts := strings.Split(mount, ":")

	// Validate we have the right number of parts (2 or 3)
	if len(parts) < 2 {
		return engine.VolumeMount{}, fmt.Errorf("invalid mount string format: must contain at least source and destination separated by ':'")
	}

	if len(parts) > 3 {
		return engine.VolumeMount{}, fmt.Errorf("invalid mount string format: too many ':' separators (expected format: source:destination[:options])")
	}

	// Extract source and destination
	source := parts[0]
	destination := parts[1]

	// Validate source is not empty
	if source == "" {
		return engine.VolumeMount{}, fmt.Errorf("source path cannot be empty")
	}

	// Validate destination is not empty
	if destination == "" {
		return engine.VolumeMount{}, fmt.Errorf("destination path cannot be empty")
	}

	vol := engine.VolumeMount{
		Source:      source,
		Destination: destination,
	}
	// Extract mount options if present
	if len(parts) == 3 {
		vol.MountOptions = parts[2]
	}

	return vol, nil
}
