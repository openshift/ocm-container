package engine

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
)

var (
	SupportedEngines = []string{"podman", "docker"}
)

type Container struct {
	ID  string
	Ref ContainerRef
}

type ContainerRef struct {
	Image       string
	Tag         string
	Volumes     []VolumeMount
	Envs        map[string]string
	Tty         bool
	Interactive bool
	Entrypoint  string
	Command     string
}

type VolumeMount struct {
	Source       string
	Destination  string
	MountOptions string
}

type Engine struct {
	engine  string
	binary  string
	verbose bool
}

func New(engine string, verbose bool) (*Engine, error) {
	e := &Engine{
		verbose: verbose,
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

// exec runs a command with args for a given container engine and prints the output
func (e *Engine) exec(subcommand string, args ...string) (string, error) {
	var err error

	command := e.engine
	args = append([]string{subcommand}, args...)

	c := exec.Command(command, args...)

	if e.verbose {
		fmt.Printf("executing command: %+v\n", c)
	}

	// stdOut is the pipe for command output
	// TODO: How do we stream this live?
	var stdOut io.ReadCloser
	stdOut, err = c.StdoutPipe()
	if err != nil {
		return "", err
	}

	// stderr is the pipe for err output
	var stdErr io.ReadCloser
	stdErr, err = c.StderrPipe()
	if err != nil {
		return "", err
	}

	err = c.Start()
	if err != nil {
		return "", err
	}

	var out []byte
	out, err = io.ReadAll(stdOut)
	if err != nil {
		return "", err
	}

	var errOut []byte
	errOut, err = io.ReadAll(stdErr)
	if err != nil {
		return "", err
	}

	fmt.Fprint(os.Stderr, string(errOut))
	return string(out), nil
}

func (e *Engine) Copy(args ...string) (string, error) {
	return e.exec("cp", args...)
}

// Exec creates a container with the given args, returning a *Container object
func (e *Engine) Create(c ContainerRef) (*Container, error) {
	err := validateContainerRef(c)
	if err != nil {
		return nil, err
	}

	args, err := parseRefToArgs(c)
	if err != nil {
		return nil, err
	}

	if e.verbose {
		fmt.Printf("creating container with args: %v\n", args)
	}

	id, err := e.exec("create", args...)
	id = strings.TrimSuffix(id, "\n")
	if err != nil {
		return nil, err
	}

	return &Container{ID: id, Ref: c}, nil
}

// Start starts a given container
func (e *Engine) Start(c *Container) error {
	var args = []string{"start"}
	args = append(args, c.ID)

	err := e.execAndReplace(args...)
	return err
}

func (e *Engine) execAndReplace(args ...string) error {
	if e.verbose {
		fmt.Printf("executing command, replacing this process: %v %v\n", e.binary, append([]string{e.engine}, args...))
	}

	// This append of the engine is correct - the first argument is also the program name
	return syscall.Exec(e.binary, append([]string{e.engine}, args...), os.Environ())
}

// Attach attaches to a container with the given id, replacing this process
func (e *Engine) Attach(c *Container) error {
	return e.execAndReplace("attach", c.ID)
}

// StartAndAttach starts a given container and attaches to it, replacing this process
func (e *Engine) StartAndAttach(c *Container) error {
	var args = []string{"start"}
	args = append(args, "--attach")
	args = append(args, c.ID)

	err := e.execAndReplace(args...)
	return err
}

// Run launches a container with the given args, replacing this process
func (e *Engine) Run(args ...string) error {
	return e.execAndReplace("run", strings.Join(args, " "))
}

// Version returns the version of the container engine, replacing this process
func (e *Engine) Version() error {
	return e.execAndReplace("version")
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

	args := []string{"--privileged"}

	if c.Envs != nil {
		args = append(args, envsToString(c.Envs)...)
	}

	if c.Volumes != nil {
		for _, v := range c.Volumes {
			args = append(args, fmt.Sprintf("--volume=%s:%s:%s", v.Source, v.Destination, v.MountOptions))
		}
	}

	if c.Entrypoint != "" {
		args = append(args, fmt.Sprintf("--entrypoint=%s", c.Entrypoint))
	}

	args = append(args, ttyToString(c.Tty, c.Interactive)...)

	args = append(args, c.Image+":"+c.Tag)

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
		args = append(args, "--env")
		args = append(args, fmt.Sprintf("%s=%s", k, v))
	}
	return args
}
