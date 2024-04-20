package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/openshift/ocm-container/pkg/subprocess"
)

var (
	SupportedEngines           = []string{"podman", "docker"}
	SupportedPullImagePolicies = []string{"always", "missing", "never", "newer"}
)

type Container struct {
	ID  string
	Ref ContainerRef
}

type ContainerImage struct {
	Registry   string
	Repository string
	Name       string
	Tag        string
	Hash       string
}

type ContainerRef struct {
	Image           ContainerImage
	Tag             string
	Volumes         []VolumeMount
	Envs            map[string]string
	Tty             bool
	PublishAll      bool
	Interactive     bool
	Entrypoint      string
	Command         string
	BestEffortArgs  []string
	Privileged      bool
	RemoveAfterExit bool
}

type VolumeMount struct {
	Source       string
	Destination  string
	MountOptions string
}

type Engine struct {
	engine     string
	binary     string
	pullPolicy string
	dryRun     bool
	verbose    bool
}

func New(engine, pullPolicy string, dryRun, verbose bool) (*Engine, error) {
	e := &Engine{
		pullPolicy: pullPolicy,
		dryRun:     dryRun,
		verbose:    verbose,
	}

	if verbose {
		fmt.Println("using container engine:", engine)
	}

	if !slices.Contains(SupportedEngines, engine) {
		err := fmt.Errorf("error: engine %s not in supported engines: %v", engine, strings.Join(SupportedEngines, ", "))
		return nil, err
	}

	bin, err := exec.LookPath(engine)
	if err != nil {
		err = fmt.Errorf("error: engine not found in $PATH: %v", err)
		return nil, err
	}

	e.engine = engine
	e.binary = bin

	return e, nil
}

// Attach attaches to a container with the given id, replacing this process
func (e *Engine) Attach(c *Container) error {
	return e.execAndReplace([]string{"attach", c.ID}...)
}

// Copy copies a source file to a destination (eg: podman cp)
func (e *Engine) Copy(cpArgs ...string) (string, error) {
	var args = []string{"cp"}
	args = append(args, cpArgs...)
	return e.exec(args...)
}

// Exec creates a container with the given args, returning a *Container object
func (e *Engine) Create(c ContainerRef) (*Container, error) {
	var err error
	var args = []string{"create", pullPolicyString(e.pullPolicy)}
	err = validateContainerRef(c)
	if err != nil {
		return nil, err
	}

	refArgs, err := parseRefToArgs(c)
	if err != nil {
		return nil, err
	}

	args = append(args, refArgs...)

	// Run the command
	id, err := e.exec(args...)

	// Trim newlines the response from the engine
	id = strings.TrimSuffix(id, "\n")
	if err != nil {
		return nil, err
	}

	return &Container{ID: id, Ref: c}, nil
}

// Exec runs a command inside a running container (eg: podman exec)
func (e *Engine) Exec(c *Container, execArgs []string) (string, error) {
	var err error
	var args = []string{"exec"}

	// The container may be --privileged, but Exec doesn't use that flag by default
	if c.Ref.Privileged {
		args = append(args, "--privileged")
	}

	args = append(args, c.ID)
	args = append(args, execArgs...)

	if e.verbose && !e.dryRun {
		fmt.Printf("executing command inside the running container: %v %v\n", e.binary, append([]string{e.engine}, args...))
	}

	out, err := e.exec(args...)

	return out, err
}

// Inspect takes a string value as a formatter for inspect output
// (eg: podman inspect --format=)
func (e *Engine) Inspect(c *Container, value string) (string, error) {
	return e.exec([]string{"inspect", c.ID, fmt.Sprintf("--format='%s'", value)}...)
}

// Start starts a given container
// (eg: podman start)
func (e *Engine) Start(c *Container, attach bool) error {
	var err error

	if attach {
		// This error won't actually be used, since the process is replaced
		_ = e.execAndReplace("start", "--attach", c.ID)
	}

	out, err := e.exec("start", c.ID)

	if e.verbose {
		fmt.Println(out)
	}

	return err
}

// Version returns the version of the container engine, replacing this process
func (e *Engine) Version() error {
	return e.execAndReplace("version")
}

// exec runs a command with args for a given container engine and prints the output
func (e *Engine) exec(args ...string) (string, error) {
	command := e.engine
	c := exec.Command(command, args...)

	e.printCmdIfVerbose(fmt.Sprint(c))
	if e.dryRun {
		return "", nil
	}

	return subprocess.Run(c)
}

func (e *Engine) execAndReplace(args ...string) error {

	// This append of the engine is correct - the first argument is also the program name
	execArgs := append([]string{e.engine}, args...)

	e.printCmdIfVerbose(fmt.Sprintf("%v %v", e.binary, execArgs))
	if e.dryRun {
		return nil
	}

	return subprocess.RunAndReplace(e.binary, execArgs, os.Environ())
}

// imageFQDN builds an image format string from container ref values
func (c ContainerRef) imageFQDN() string {
	i := fmt.Sprintf("%s:%s", c.Image.Name, c.Image.Tag)

	// The order of the repository and registry addition is important
	if c.Image.Repository != "" {
		i = fmt.Sprintf("%s/%s", c.Image.Repository, i)
	}

	if c.Image.Registry != "" {
		i = fmt.Sprintf("%s/%s", c.Image.Registry, i)
	}

	return i
}

// validateContainerRef tries to do some pre-validation of the ref data to avoid process errors
func validateContainerRef(c ContainerRef) error {
	for _, v := range c.Volumes {
		if v.Source == "" || v.Destination == "" {
			return fmt.Errorf("error: invalid volume mount: %v", v)
		}

		if _, err := os.Stat(v.Source); err != nil {
			return fmt.Errorf("error: problem reading source volume: %v: %v", v.Source, err)
		}

		v.Source = filepath.Clean(v.Source)
		v.Destination = filepath.Clean(v.Destination)
	}
	return nil
}

// parseRefToArgs converts a ContainerRef to a slice of strings for use in exec
func parseRefToArgs(c ContainerRef) ([]string, error) {
	var args []string

	if c.Privileged {
		args = append(args, "--privileged")
	}

	if c.RemoveAfterExit {
		args = append(args, "--rm")
	}

	if c.PublishAll {
		args = append(args, "--publish-all")
	}

	if c.Envs != nil {
		args = append(args, envsToString(c.Envs)...)
	}

	if c.Volumes != nil {
		for _, v := range c.Volumes {
			args = append(args, fmt.Sprintf("--volume=%s:%s:%s", v.Source, v.Destination, v.MountOptions))
		}
	}

	if c.BestEffortArgs != nil {
		args = append(args, c.BestEffortArgs...)
	}

	if c.Entrypoint != "" {
		args = append(args, fmt.Sprintf("--entrypoint=%s", c.Entrypoint))
	}

	args = append(args, ttyToString(c.Tty, c.Interactive)...)

	args = append(args, c.imageFQDN())

	// This needs to come last because command is a positional argument
	if c.Command != "" {
		args = append(args, c.Command)
	}

	return args, nil
}

// tty converts the tty and interactive bool values to string cli args
func ttyToString(tty, interactive bool) []string {
	var args []string

	if tty {
		args = append(args, "--tty")
	}

	if interactive && tty {
		args = append(args, "--interactive")
	}

	return args
}

// envsToString converts a map[string]string of envs to a slice of strings for use in exec
func envsToString(envs map[string]string) []string {
	var args []string
	for k, v := range envs {
		// Any spaces (eg: between '--env' and the key/value pair) MUST be
		// appended as individual strings to the slice, not as a single string
		// Boo: []string{"--env key=value"}; Yay: []string{"--env", "key=value"}
		args = append(args, "--env")
		if v == "" {
			args = append(args, k)
		} else {
			args = append(args, fmt.Sprintf("%s=%s", k, v))
		}
	}
	return args
}

// --quiet suppresses image pull output which is written to /dev/null and
// misinterpreted by os.Exec as an error message
func pullPolicyString(s string) string {
	return fmt.Sprintf("--pull=%s", s)
}

func (e *Engine) printCmdIfVerbose(c string) {
	if e.verbose && !e.dryRun {
		fmt.Printf("executing command: %+v\n", c)
	}

	if e.dryRun {
		fmt.Printf("dry-run; would have executed: %+v\n", c)
	}
}
