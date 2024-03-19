package ocmcontainer

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/ocm-container/pkg/backplane"
	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/openshift/ocm-container/pkg/gcloud"
	"github.com/openshift/ocm-container/pkg/jira"
	"github.com/openshift/ocm-container/pkg/osdctl"
	"github.com/openshift/ocm-container/pkg/pagerduty"
	"github.com/spf13/cobra"
)

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

	// image, tag, launchOpts, console, personalization
	image, tag, _, _, _, err := parseFlags(cmd)
	if err != nil {
		return o, err
	}

	c := engine.ContainerRef{
		Image:       image,
		Tag:         tag,
		Tty:         true,
		Interactive: true,
	}

	c.Envs = make(map[string]string)

	// Standard OCM container user environment envs
	// Setting the strings to empty will pass them in
	// in the "-e USER" from the environment format
	c.Envs["OCM_URL"] = ""
	c.Envs["USER"] = ""
	c.Envs["OFFLINE_ACCESS_TOKEN"] = ""

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
	c.Volumes = append(c.Volumes, backplaneConfig.Mount)

	// PagerDuty configuration
	pagerDutyConfig, err := pagerduty.New(home)
	if err != nil {
		return o, err
	}
	c.Volumes = append(c.Volumes, pagerDutyConfig.Mount)

	// Jira configuration
	jiraConfig, err := jira.New(home)
	if err != nil {
		return o, err
	}
	maps.Copy(jiraConfig.Env, c.Envs)
	c.Volumes = append(c.Volumes, jiraConfig.Mount)

	// OSDCTL configuration
	osdctlConfig, err := osdctl.New(home)
	if err != nil {
		return o, err
	}
	c.Volumes = append(c.Volumes, osdctlConfig.Mount)

	// GCloud configuration
	gcloudConfig, err := gcloud.New(home)
	if err != nil {
		return o, err
	}
	c.Volumes = append(c.Volumes, gcloudConfig.Mount)

	cluster, command, err := parseArgs(args)
	if err != nil {
		return o, err
	}

	if cluster != "" {
		if o.verbose {
			fmt.Printf("logging into cluster: %s\n", cluster)
		}
		c.Envs["INITIAL_CLUSTER_LOGIN"] = cluster
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
func parseFlags(cmd *cobra.Command) (image, tag, launchOpts string, console, personalization bool, err error) {
	image, err = cmd.Flags().GetString("image")
	if err != nil {
		return "", "", "", false, false, err
	}

	tag, err = cmd.Flags().GetString("tag")
	if err != nil {
		return "", "", "", false, false, err
	}

	launchOpts, err = cmd.Flags().GetString("launch-opts")
	if err != nil {
		return "", "", "", false, false, err
	}

	// We're swapping the negative "disable-console-port" for a positive variable name
	// disable-console-port is negative with default value "false" (double negative), so we're setting console to true
	c, err := cmd.Flags().GetBool("disable-console-port")
	if err != nil {
		return "", "", "", false, false, err
	}
	console = !c

	// We're swapping the negative "no-personalization" for a positive variable name
	// no-personalization is negative with default value "true" so we're setting personalization to false
	p, err := cmd.Flags().GetBool("no-personalizations")
	if err != nil {
		return "", "", "", false, false, err
	}
	personalization = !p

	return image, tag, launchOpts, console, personalization, err
}

// parseArgs takes a slice of strings and returns the clusterID and the command to execute inside the container
func parseArgs(args []string) (string, string, error) {
	switch {
	case len(args) == 1:
		return args[0], "", nil
	case len(args) > 1:
		// TODO: I don't understand why "--" is not parsed as an argument here, and disappears from the []string
		// We definitely want to try to make this work if we can
		// if args[1] != "--" {
		// 	e := strings.Builder{}
		// 	e.WriteString(fmt.Sprintf("invalid arguments: %s; expected format: \n", args[1]))
		// 	e.WriteString("\tocm-container [FLAGS] <clusterID> -- <command>\n")
		// 	e.WriteString("\tocm-container [FLAGS] <clusterID>\n")
		// 	return "", "", errors.New(e.String())
		// }

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

func (o *ocmContainer) Exec() error {
	// Not yet implemented
	return nil
}

func (o *ocmContainer) Copy(source, destination string) error {
	s := filepath.Clean(source)
	d := filepath.Clean(destination)

	args := fmt.Sprintf("%s:%s", s, d)

	o.engine.Copy("cp", args)

	return nil
}
