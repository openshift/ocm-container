package engine

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"syscall"
)

var (
	SupportedEngines = []string{"podman", "docker"}
)

type Container struct {
	ID string
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
		fmt.Println("Using container engine:", engine)
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
func (e *Engine) Create(args ...string) (*Container, error) {
	id, err := e.exec("create", args...)
	id = strings.TrimSuffix(id, "\n")
	if err != nil {
		return nil, err
	}

	return &Container{ID: id}, nil
}

// Start starts a given container
func (e *Engine) Start(c *Container) error {
	_, err := e.exec("start", c.ID)
	return err
}

func (e *Engine) execAndReplace(args ...string) error {
	return syscall.Exec(e.binary, append([]string{e.engine}, args...), os.Environ())
}

// Attach attaches to a container with the given id, replacing this process
func (e *Engine) Attach(c *Container) error {
	return e.execAndReplace("attach", c.ID)

}

// StartAndAttach starts a given container and attaches to it, replacing this process
func (e *Engine) StartAndAttach(c *Container) error {
	err := e.Start(c)
	if err != nil {
		return err
	}
	return e.Attach(c)
}

// Run launches a container with the given args, replacing this process
func (e *Engine) Run(args ...string) error {
	return e.execAndReplace("run", strings.Join(args, " "))
}

// Version returns the version of the container engine, replacing this process
func (e *Engine) Version() error {
	return e.execAndReplace("version")
}
