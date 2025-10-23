package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/openshift/ocm-container/pkg/subprocess"
	log "github.com/sirupsen/logrus"
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
	Image   ContainerImage
	Tag     string
	Volumes []VolumeMount
	Envs    []EnvVar
	// Retain this for now for backwards compatibility
	// until we have all features migrated, then we
	// will remove the EnvMap
	EnvMap          map[string]string
	Tty             bool
	PublishAll      bool
	Interactive     bool
	Entrypoint      string
	Command         string
	BestEffortArgs  []string
	Privileged      bool
	RemoveAfterExit bool
	LocalPorts      map[string]int
}

type VolumeMount struct {
	Source       string
	Destination  string
	MountOptions string
}

type EnvVar struct {
	Key   string
	Value string
}

func (e *EnvVar) Parse() (string, error) {
	if e.Key == "" {
		return "", fmt.Errorf("env key not present")
	}
	if e.Value == "" {
		return fmt.Sprintf("-e %s", e.Key), nil
	}
	return fmt.Sprintf("-e %s=%s", e.Key, e.Value), nil
}

func EnvVarFromString(str string) (EnvVar, error) {
	// TODO - can we do this validation using the same functions as podman?
	e := EnvVar{}
	if str == "" {
		return e, fmt.Errorf("unexpected empty string for env")
	}
	kv := strings.Split(str, "=")
	if len(kv) == 0 {
		return e, fmt.Errorf("Unexpected empty split for env: %s", str)
	}
	if len(kv) > 2 {
		return e, fmt.Errorf("Length of env string split > 2 for env: %s", str)
	}
	if len(kv) == 2 {
		e.Value = kv[1]
	}
	e.Key = kv[0]
	return e, nil
}

type Engine struct {
	engine     string
	binary     string
	pullPolicy string
	dryRun     bool
}

func New(engine, pullPolicy string, dryRun bool) (*Engine, error) {
	e := &Engine{
		pullPolicy: pullPolicy,
		dryRun:     dryRun,
	}

	log.Debug(fmt.Sprintf("using container engine: %s", engine))

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
	log.Debugf("executing command to copy files: %v %v\n", e.binary, args)
	return e.exec(args...)
}

// Exec creates a container with the given args, returning a *Container object
func (e *Engine) Create(c ContainerRef) (*Container, error) {
	var err error
	var args = []string{"create", pullPolicyToString(e.pullPolicy)}

	// --quiet suppresses image pull policy output which is written to /dev/null and
	// misinterpreted by os.Exec as an error message
	switch log.GetLevel() {
	case log.TraceLevel:
		// pass
	case log.DebugLevel:
		// pass
	default:
		args = append(args, "--quiet")
	}

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

	if !e.dryRun {
		log.Debugf("executing command inside the running container: %v %v\n", e.binary, args)
	}

	out, err := e.exec(args...)

	return out, err
}

func (e *Engine) ExecLive(c *Container, execArgs []string) error {
	var err error
	var args = []string{"exec", "--interactive", "--tty"}

	// The container may be --privileged, but Exec doesn't use that flag by default
	if c.Ref.Privileged {
		args = append(args, "--privileged")
	}

	args = append(args, c.ID)
	args = append(args, execArgs...)

	if !e.dryRun {
		log.Debugf("executing command inside the running container: %v %v\n", e.binary, args)
	}

	// we don't care about output here since that will get forwarded to the terminal interactively
	_, err = e.run(args...)

	return err
}

// Inspect takes a string value as a formatter for inspect output
// (eg: podman inspect --format=)
func (e *Engine) Inspect(c *Container, value string) (string, error) {
	return e.exec([]string{"inspect", c.ID, fmt.Sprintf("--format='%s'", value)}...)
}

func (e *Engine) Stop(c *Container, timeout int) error {
	_, err := e.exec([]string{"stop", c.ID, fmt.Sprintf("--time=%d", timeout)}...)
	return err
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

	// This is not log output; do not pass through a logger
	log.Debug("Exec output: " + out)

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

	return subprocess.Run(c)
}

func (e *Engine) run(args ...string) (string, error) {
	command := e.engine
	c := exec.Command(command, args...)
	return subprocess.RunLive(c)
}

func (e *Engine) execAndReplace(args ...string) error {
	// This append of the engine is correct - the first argument is also the program name
	execArgs := append([]string{e.engine}, args...)
	return subprocess.RunAndReplace(e.binary, execArgs, os.Environ())
}

// imageFQDN builds an image format string from container ref values
func (c ContainerRef) imageFQDN() string {
	if c.Image.Name == "" || c.Image.Tag == "" {
		return ""
	}

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

	// If PublishAll is set - publish all of the ports. Otherwise, bind each
	// one individually to localhost.
	if c.PublishAll {
		args = append(args, "--publish-all")
	} else if c.LocalPorts != nil {
		for service := range c.LocalPorts {
			args = append(args, fmt.Sprintf("--publish=127.0.0.1::%d", c.LocalPorts[service]))
		}
	}

	if c.EnvMap != nil {
		args = append(args, envMapToString(c.EnvMap)...)
	}

	if c.Envs != nil {
		args = append(args, envsToString(c.Envs)...)
	}

	if c.Volumes != nil {
		args = append(args, volumesToString(c.Volumes)...)
	}

	if c.BestEffortArgs != nil {
		args = append(args, c.BestEffortArgs...)
	}

	if c.Entrypoint != "" {
		args = append(args, fmt.Sprintf("--entrypoint=%s", c.Entrypoint))
	}

	args = append(args, ttyToString(c.Tty, c.Interactive)...)

	imageFQDN := c.imageFQDN()
	if imageFQDN != "" {
		args = append(args, imageFQDN)
	}

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

func envsToString(envs []EnvVar) []string {
	var args []string
	for k := range envs {
		val, err := envs[k].Parse()
		if err != nil {
			log.Warnf("error parsing environment variables: %v", err)
		}
		args = append(args, val)
	}

	return args
}

// envMapToString converts a map[string]string of envs to a slice of strings for use in exec
func envMapToString(envs map[string]string) []string {
	var args []string

	keys := make([]string, 0, len(envs))
	for k := range envs {
		keys = append(keys, k)
	}

	// Sorting the keys ensures that the order of environment variables is consistent, for testing and reproducibility
	sort.Strings(keys)

	for _, k := range keys {
		// Any spaces (eg: between '--env' and the key/value pair) MUST be
		// appended as individual strings to the slice, not as a single string
		// Boo: []string{"--env key=value"}; Yay: []string{"--env", "key=value"}

		args = append(args, "--env")

		if envs[k] == "" {
			args = append(args, k)
		} else {
			args = append(args, fmt.Sprintf("%s=%s", k, envs[k]))
		}
	}
	return args
}

func volumesToString(volumes []VolumeMount) []string {
	args := []string{}
	for _, v := range volumes {
		mountString := fmt.Sprintf("%s:%s", v.Source, v.Destination)
		if v.MountOptions != "" {
			mountString = mountString + ":" + v.MountOptions
		}
		args = append(args, fmt.Sprintf("--volume=%s", mountString))
	}
	return args
}

// pullPolicyToString returns a string for the --pull flag as a valid container engine argument
func pullPolicyToString(s string) string {
	return fmt.Sprintf("--pull=%s", s)
}
