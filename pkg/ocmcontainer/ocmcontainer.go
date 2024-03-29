package ocmcontainer

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/ocm-container/pkg/aws"
	"github.com/openshift/ocm-container/pkg/backplane"
	"github.com/openshift/ocm-container/pkg/certificateAuthorities"
	"github.com/openshift/ocm-container/pkg/deprecation"
	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/openshift/ocm-container/pkg/gcloud"
	"github.com/openshift/ocm-container/pkg/jira"
	"github.com/openshift/ocm-container/pkg/opsutils"
	"github.com/openshift/ocm-container/pkg/osdctl"
	"github.com/openshift/ocm-container/pkg/pagerduty"
	personalize "github.com/openshift/ocm-container/pkg/personalization"
	"github.com/openshift/ocm-container/pkg/scratch"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var errClusterAndUnderscoreArgs = errors.New("specifying a cluster with --cluster and using an underscore in the first argument are mutually exclusive")

type ocmContainer struct {
	engine    *engine.Engine
	container *engine.Container
	verbose   bool
}

func New(cmd *cobra.Command, args []string, containerEngine string, dryRun, verbose bool) (*ocmContainer, error) {
	var err error

	o := &ocmContainer{
		verbose: verbose,
	}

	o.engine, err = engine.New(containerEngine, dryRun, verbose)
	if err != nil {
		return o, err
	}

	c := engine.ContainerRef{}
	// image, tag, launchOpts, console, personalization
	c, err = parseFlags(cmd, c)
	if err != nil {
		return o, err
	}
	if verbose {
		fmt.Printf("container ref: %+v\n", c)
	}

	// We're swapping the negative "disable-console-port" for a positive variable name
	// disable-console-port is negative with default value "false" (double negative), so we're setting console to true
	console, err := cmd.Flags().GetBool("disable-console-port")
	if err != nil {
		return o, err
	}
	console = !console
	c.PublishAll = console

	// TODO: PARSE LAUNCH OPTS SOMEWHERE
	// launchOpts, err := cmd.Flags().GetString("launch-opts")
	// if err != nil {
	// 	return c, err
	// }

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

	// PagerDuty configuration
	pagerDutyConfig, err := pagerduty.New(home)
	if err != nil {
		return o, err
	}
	c.Volumes = append(c.Volumes, pagerDutyConfig.Mounts...)

	// Optional Certificate Authority Trust mount
	if viper.IsSet("ca_source_anchors") {
		caConfig, err := certificateAuthorities.New(viper.GetString("ca_source_anchors"))
		if err != nil {
			return o, err
		}
		c.Volumes = append(c.Volumes, caConfig.Mounts...)
	}

	// Jira configuration
	jiraConfig, err := jira.New(home)
	if err != nil {
		return o, err
	}
	maps.Copy(jiraConfig.Env, c.Envs)
	c.Volumes = append(c.Volumes, jiraConfig.Mounts...)

	// OSDCTL configuration
	osdctlConfig, err := osdctl.New(home)
	if err != nil {
		return o, err
	}
	c.Volumes = append(c.Volumes, osdctlConfig.Mounts...)

	// AWS Credentials
	awsConfig, err := aws.New(home)
	if err != nil {
		return o, err
	}
	c.Volumes = append(c.Volumes, awsConfig.Mounts...)

	// GCloud configuration
	gcloudConfig, err := gcloud.New(home)
	if err != nil {
		return o, err
	}
	c.Volumes = append(c.Volumes, gcloudConfig.Mounts...)

	// CA Trust Source Anchor mount
	ca_trust_source_anchors := viper.GetString("ca_source_anchors")
	if ca_trust_source_anchors != "" {
		caAnchorsConfig, err := certificateAuthorities.New(ca_trust_source_anchors)
		if err != nil {
			return o, err
		}
		c.Volumes = append(c.Volumes, caAnchorsConfig.Mounts...)
	}

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

	// Scratch Dir mount

	scratchDir := viper.GetString("scratch_dir")
	scratchDirRWMount := viper.GetBool("scratch_dir_rw")
	if scratchDir != "" {
		scratchConfig, err := scratch.New(scratchDir, scratchDirRWMount)
		if err != nil {
			return o, err
		}
		c.Volumes = append(c.Volumes, scratchConfig.Mounts...)
	}

	// Personalization
	personalization := viper.GetBool("enable_personalization_mount")
	personalizationDirOrFile := viper.GetString("personalization_file")
	personalizationRWMount := viper.GetBool("scratch_dir_rw")

	if personalization && personalizationDirOrFile != "" {
		personalizationConfig, err := personalize.New(personalizationDirOrFile, personalizationRWMount)
		if err != nil {
			return o, err
		}
		c.Volumes = append(c.Volumes, personalizationConfig.Mounts...)
	}

	// Future-proofing this: if -c is provided for a cluster ID instead of a positional argument,
	// then parseArgs should just treat all positional arguments as the command to run in the container

	cluster, err := cmd.Flags().GetString("cluster")
	if err != nil {
		return o, err
	}

	cluster, command, err := parseArgs(args, cluster)
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

// parseFlags takes a cobra command and returns the flags as strings or bool values
func parseFlags(cmd *cobra.Command, c engine.ContainerRef) (engine.ContainerRef, error) {

	c.Tty = true
	c.Interactive = true

	entrypoint, err := cmd.Flags().GetString("entrypoint")
	if err != nil {
		return c, err
	}
	if entrypoint != "" {
		c.Entrypoint = entrypoint
	}

	// This is a deprecated command - the same can be accomplished with engine-specific
	// entrypoint and positional CMD arguments - but we're keeping it for now to socialize it
	exec, err := cmd.Flags().GetString("exec")
	if err != nil {
		return c, err
	}
	if exec != "" {
		deprecation.Message("--exec", "--entrypoint")
		c.Command = exec
		c.Tty = false
		c.Interactive = false
	}

	// Image options

	registry, err := cmd.Flags().GetString("registry")
	if err != nil {
		return c, err
	}

	repository, err := cmd.Flags().GetString("repository")
	if err != nil {
		return c, err
	}

	image, err := cmd.Flags().GetString("image")
	if err != nil {
		return c, err
	}

	tag, err := cmd.Flags().GetString("tag")
	if err != nil {
		return c, err
	}

	i := engine.ContainerImage{
		Registry:   registry,
		Repository: repository,
		Name:       image,
		Tag:        tag,
	}

	c.Image = i

	return c, err
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
		deprecation.Message("using cluster ids in a positional argument", "--cluster")
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

func (o *ocmContainer) Start(attach bool) error {
	if attach {
		return o.engine.StartAndAttach(o.container)
	}

	return o.engine.Start(o.container)
}

func (o *ocmContainer) StartAndAttach() error {
	err := o.Start(true)
	if err != nil {
		return err
	}
	return o.engine.Attach(o.container)
}

func (o *ocmContainer) Exec(args []string) error {
	return o.engine.Exec(o.container, args)
}

func (o *ocmContainer) Copy(source, destination string) error {
	s := filepath.Clean(source)
	d := filepath.Clean(destination)

	args := fmt.Sprintf("%s:%s", s, d)

	o.engine.Copy("cp", args)

	return nil
}
