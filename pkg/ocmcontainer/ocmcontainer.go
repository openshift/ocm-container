package ocmcontainer

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/spf13/cobra"
)

type containerRef struct {
	image       string
	tag         string
	volumes     []volumeMount
	envs        map[string]string
	tty         bool
	interactive bool
	entrypoint  string
}

type volumeMount struct {
	source       string
	destination  string
	mountOptions string
}

type ocmContainer struct {
	engine    *engine.Engine
	container *engine.Container
	cluster   string
	verbose   bool
}

func New(cmd *cobra.Command, args []string, containerEngine string, verbose bool) (*ocmContainer, error) {
	var err error

	o := &ocmContainer{
		verbose: verbose,
	}

	o.engine, err = engine.New(containerEngine, verbose)
	if err != nil {
		return o, err
	}

	// image, tag, launchOpts, console, personalization
	if o.verbose {
		fmt.Println("parsing flags")
	}

	image, tag, _, _, _, err := parseFlags(cmd)
	if err != nil {
		return o, err
	}

	c := containerRef{
		image:       image,
		tag:         tag,
		tty:         true,
		interactive: true,
	}

	// Parse the initial cluster login and entrypoint from the CLI args; if any
	if o.verbose {
		fmt.Println("parsing arguments")
	}
	cluster, entrypoint, err := parseArgs(args)
	if err != nil {
		return o, err
	}

	if cluster != "" {
		if o.verbose {
			fmt.Printf("logging into cluster: %s\n", cluster)
		}
		c.envs["INITIAL_CLUSTER_LOGIN"] = cluster
	}

	if entrypoint != "" {
		if o.verbose {
			fmt.Printf("setting entrypoint: %s\n", entrypoint)
		}
		c.entrypoint = entrypoint
	}

	// Create the actual container
	if o.verbose {
		fmt.Printf("using containerRef: %v\n", c)
	}
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
		if args[1] != "--" {
			e := strings.Builder{}
			e.WriteString(fmt.Sprintf("invalid arguments: %s; expected format: \n", args[1]))
			e.WriteString("\tocm-container [FLAGS] <clusterID> -- <command>\n")
			e.WriteString("\tocm-container [FLAGS] <clusterID>\n")
			return "", "", errors.New(e.String())
		}
		return args[0], strings.Join(args[2:], " "), nil
	}
	return "", "", nil
}

// This is just a wrapper around Create to unpack the containerRef
func (o *ocmContainer) CreateContainer(c containerRef) error {
	var args []string

	args = append(args, tty(c.tty, c.interactive)...)
	args = append(args, c.image+":"+c.tag)

	return o.Create(args...)
}

func (o *ocmContainer) Create(args ...string) error {
	if o.verbose {
		fmt.Printf("creating container with args: %+v\n", strings.Join(args, ", "))
	}
	container, err := o.engine.Create(args...)
	if err != nil {
		return err
	}
	o.container = container
	return nil
}

func (o *ocmContainer) Start() error {
	return o.engine.Start(o.container)
}

func (o *ocmContainer) StartAndAttach() error {
	err := o.Start()
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

func tty(tty, interactive bool) []string {
	var args []string

	if tty {
		args = append(args, "--tty")
	}

	if interactive && tty {
		args = append(args, "--interactive")
	}

	return args
}
