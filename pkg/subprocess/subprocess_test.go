package subprocess

import (
	"fmt"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

// testHelperProcess is the re-exec entry point. When the test binary is
// invoked with GO_TEST_HELPER_PROCESS=1, it acts as a fake subprocess
// instead of running tests. This avoids shelling out to host commands.
func init() {
	if os.Getenv("GO_TEST_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args
	for i, arg := range args {
		if arg == "--" {
			args = args[i+1:]
			break
		}
	}

	if len(args) == 0 {
		os.Exit(1)
	}

	switch args[0] {
	case "echo":
		fmt.Println(args[1])
	case "print-no-newline":
		fmt.Print(args[1])
	case "exit-nonzero":
		fmt.Fprintln(os.Stderr, "something failed")
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "unknown helper command: %s\n", args[0])
		os.Exit(2)
	}

	os.Exit(0)
}

// helperCommand builds an exec.Cmd that re-invokes the test binary as a
// fake subprocess. The helper dispatches on args[0] after "--".
func helperCommand(command string, args ...string) *exec.Cmd {
	allArgs := []string{"-test.run=^$", "--", command}
	allArgs = append(allArgs, args...)
	cmd := exec.Command(os.Args[0], allArgs...)
	cmd.Env = append(os.Environ(), "GO_TEST_HELPER_PROCESS=1")
	return cmd
}

var _ = Describe("Pkg/Subprocess/Subprocess", func() {
	BeforeEach(func() {
		viper.Reset()
	})

	Context("ExecErr", func() {
		It("Strips 'Error: ' prefix from stderr", func() {
			ee := &ExecErr{
				ExecStdErr: "Error: something went wrong\n",
			}
			Expect(ee.Error()).To(Equal("something went wrong"))
		})

		It("Strips 'error: ' lowercase prefix from stderr", func() {
			ee := &ExecErr{
				ExecStdErr: "error: something went wrong\n",
			}
			Expect(ee.Error()).To(Equal("something went wrong"))
		})

		It("Strips trailing newlines from stderr", func() {
			ee := &ExecErr{
				ExecStdErr: "some error\n",
			}
			Expect(ee.Error()).To(Equal("some error"))
		})

		It("Returns unmodified stderr when no prefix or trailing newline", func() {
			ee := &ExecErr{
				ExecStdErr: "plain error",
			}
			Expect(ee.Error()).To(Equal("plain error"))
		})
	})

	Context("dryRun()", func() {
		It("Returns false when dry-run is not set", func() {
			Expect(dryRun()).To(BeFalse())
		})

		It("Returns true when dry-run is set", func() {
			viper.Set("dry-run", true)
			Expect(dryRun()).To(BeTrue())
		})
	})

	Context("Run()", func() {
		It("Captures stdout from a helper process", func() {
			cmd := helperCommand("echo", "hello")
			out, err := Run(cmd)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal("hello\n"))
		})

		It("Returns empty string in dry-run mode without executing", func() {
			viper.Set("dry-run", true)
			cmd := helperCommand("echo", "hello")
			out, err := Run(cmd)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(BeEmpty())
		})

		It("Returns ExecErr for a failing helper process", func() {
			cmd := helperCommand("exit-nonzero")
			_, err := Run(cmd)
			Expect(err).To(HaveOccurred())
			execErr, ok := err.(*ExecErr)
			Expect(ok).To(BeTrue())
			Expect(execErr.Code()).To(Equal(1))
		})

		It("Captures stdout without trailing newline from helper", func() {
			cmd := helperCommand("print-no-newline", "line1\nline2")
			out, err := Run(cmd)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal("line1\nline2"))
		})
	})

	Context("RunLive()", func() {
		It("Captures stdout from a helper process", func() {
			cmd := helperCommand("echo", "live-output")
			out, err := RunLive(cmd)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(ContainSubstring("live-output"))
		})

		It("Returns error for a failing helper process", func() {
			cmd := helperCommand("exit-nonzero")
			_, err := RunLive(cmd)
			Expect(err).To(HaveOccurred())
		})
	})
})
