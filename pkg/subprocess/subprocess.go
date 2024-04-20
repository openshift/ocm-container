package subprocess

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/spf13/viper"
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

	printCmdIfVerbose(fmt.Sprintf("%s %s %s", strings.Join(env, " "), command, strings.Join(args, " ")))
	if dryRun() {
		return nil
	}

	return syscall.Exec(command, args, env)
}

func RunLive(c *exec.Cmd) (string, error) {
	var stdoutBuf, stderrBuf bytes.Buffer
	c.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	c.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err := c.Run()
	if err != nil {
		return "", err
	}

	// TODO: Handle errStr (stderrBuf), and exit codes like in Run()
	outStr, _ := stdoutBuf.String(), stderrBuf.String()
	return outStr, nil
}

func Run(c *exec.Cmd) (string, error) {
	var err error

	printCmdIfVerbose(fmt.Sprintf("%s %s", strings.Join(c.Env, " "), c))
	if dryRun() {
		return "", nil
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

func printCmdIfVerbose(c string) {
	if verbose() && !dryRun() {
		fmt.Printf("executing command: %+v\n", c)
	}

	if dryRun() {
		fmt.Printf("dry-run; would have executed: %+v\n", c)
	}
}

func verbose() bool {
	return (viper.GetBool("verbose") || viper.GetBool("debug"))
}

func dryRun() bool {
	return viper.GetBool("dry-run")
}
