package engine

import (
	"fmt"
	"strings"
)

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
