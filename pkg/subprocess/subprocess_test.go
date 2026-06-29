package subprocess

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

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
		It("Captures stdout from a simple command", func() {
			cmd := exec.Command("echo", "hello")
			out, err := Run(cmd)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal("hello\n"))
		})

		It("Returns empty string in dry-run mode", func() {
			viper.Set("dry-run", true)
			cmd := exec.Command("echo", "hello")
			out, err := Run(cmd)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(BeEmpty())
		})

		It("Returns ExecErr for a failing command", func() {
			cmd := exec.Command("false")
			_, err := Run(cmd)
			Expect(err).To(HaveOccurred())
			execErr, ok := err.(*ExecErr)
			Expect(ok).To(BeTrue())
			Expect(execErr.Code()).To(Equal(1))
		})

		It("Captures stdout from a multi-word command", func() {
			cmd := exec.Command("printf", "line1\nline2")
			out, err := Run(cmd)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal("line1\nline2"))
		})
	})

	Context("RunLive()", func() {
		It("Captures stdout from a simple command", func() {
			cmd := exec.Command("echo", "live-output")
			out, err := RunLive(cmd)
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(ContainSubstring("live-output"))
		})

		It("Returns error for a failing command", func() {
			cmd := exec.Command("false")
			_, err := RunLive(cmd)
			Expect(err).To(HaveOccurred())
		})
	})
})
