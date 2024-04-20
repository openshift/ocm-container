package subprocess

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"syscall"
)

type ExecErr struct {
	Err        error
	ExitErr    *exec.ExitError
	ExecStdErr string
}

func (ee *ExecErr) Error() string {
	// Podman will return errors with the same formatting as this program
	// so strip out `Error: ` prefix and `\n` suffixes, since ocm-container
	// will just put them back
	s := ee.ExecStdErr
	s = strings.TrimPrefix(s, "Error: ")
	s = strings.TrimPrefix(s, "error: ")
	s = strings.TrimSuffix(s, "\n")
	return s
}

func (ee *ExecErr) Code() int {
	return ee.ExitErr.ExitCode()
}

func RunAndReplace(command string, args, env []string) error {
	return syscall.Exec(command, args, env)
}

func Run(c *exec.Cmd) (string, error) {
	var err error

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

	cmdErr := c.Start()
	if cmdErr != nil {
		return "", cmdErr
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

	errOutStr := string(errOut)

	processErr := c.Wait()
	if processErr != nil {
		if exitError, ok := processErr.(*exec.ExitError); ok {
			return "", &ExecErr{
				Err:        processErr,
				ExitErr:    exitError,
				ExecStdErr: errOutStr,
			}
		}
		return "", processErr
	}

	if errOutStr != "" {
		fmt.Println(errOutStr)
	}

	return string(out), nil
}
