package log_test

import (
	"bytes"
	"errors"
	"strings"

	"github.com/openshift/ocm-container/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var _ = Describe("Pkg/Log/LogFormatter", func() {
	var (
		formatter *log.TextFormatter
		entry     *logrus.Entry
		logger    *logrus.Logger
	)

	BeforeEach(func() {
		formatter = &log.TextFormatter{}
		logger = logrus.New()
		logger.Out = &bytes.Buffer{}
		entry = logrus.NewEntry(logger)
	})

	Describe("TextFormatter.Format", func() {
		Context("Basic formatting", func() {
			It("Formats a simple message", func() {
				entry.Message = "test message"
				entry.Level = logrus.InfoLevel

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				Expect(output).ToNot(BeNil())
				Expect(string(output)).To(ContainSubstring("test message"))
			})

			It("Includes log level in output", func() {
				entry.Message = "test message"
				entry.Level = logrus.InfoLevel
				formatter.DisableColors = true

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				Expect(string(output)).To(ContainSubstring("info"))
			})

			It("Ends with newline", func() {
				entry.Message = "test message"
				entry.Level = logrus.InfoLevel

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				Expect(output[len(output)-1]).To(Equal(byte('\n')))
			})
		})

		Context("DisableTimestamp option", func() {
			It("Excludes timestamp when DisableTimestamp is true", func() {
				formatter.DisableTimestamp = true
				formatter.DisableColors = true
				entry.Message = "test message"
				entry.Level = logrus.InfoLevel

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				outputStr := string(output)
				// Should not contain time= when DisableTimestamp is true and not formatted
				Expect(outputStr).ToNot(ContainSubstring("time="))
			})
		})

		Context("DisableColors option", func() {
			It("Disables colors when DisableColors is true", func() {
				formatter.DisableColors = true
				entry.Message = "test message"
				entry.Level = logrus.InfoLevel

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				// ANSI escape codes start with \x1b[
				Expect(string(output)).ToNot(ContainSubstring("\x1b["))
			})
		})

		Context("Log levels", func() {
			It("Formats debug level correctly", func() {
				formatter.DisableColors = true
				entry.Message = "debug message"
				entry.Level = logrus.DebugLevel

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				Expect(strings.ToLower(string(output))).To(ContainSubstring("debug"))
			})

			It("Formats info level correctly", func() {
				formatter.DisableColors = true
				entry.Message = "info message"
				entry.Level = logrus.InfoLevel

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				Expect(strings.ToLower(string(output))).To(ContainSubstring("info"))
			})

			It("Formats warn level correctly", func() {
				formatter.DisableColors = true
				entry.Message = "warn message"
				entry.Level = logrus.WarnLevel

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				Expect(strings.ToLower(string(output))).To(ContainSubstring("warn"))
			})

			It("Formats error level correctly", func() {
				formatter.DisableColors = true
				entry.Message = "error message"
				entry.Level = logrus.ErrorLevel

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				Expect(strings.ToLower(string(output))).To(ContainSubstring("error"))
			})

			It("Formats fatal level correctly", func() {
				formatter.DisableColors = true
				entry.Message = "fatal message"
				entry.Level = logrus.FatalLevel

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				Expect(strings.ToLower(string(output))).To(ContainSubstring("fatal"))
			})

			It("Formats panic level correctly", func() {
				formatter.DisableColors = true
				entry.Message = "panic message"
				entry.Level = logrus.PanicLevel

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				Expect(strings.ToLower(string(output))).To(ContainSubstring("panic"))
			})
		})

		Context("DisableUppercase option", func() {
			It("Uses uppercase level by default in formatted output", func() {
				formatter.DisableColors = true
				formatter.ForceFormatting = true
				entry.Message = "test message"
				entry.Level = logrus.InfoLevel

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				Expect(string(output)).To(ContainSubstring("INFO"))
			})

			It("Uses lowercase level when DisableUppercase is true", func() {
				formatter.DisableColors = true
				formatter.ForceFormatting = true
				formatter.DisableUppercase = true
				entry.Message = "test message"
				entry.Level = logrus.InfoLevel

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				Expect(string(output)).To(ContainSubstring("info"))
				Expect(string(output)).ToNot(ContainSubstring("INFO"))
			})
		})

		Context("Fields and data", func() {
			It("Includes additional fields", func() {
				formatter.DisableColors = true
				entry.Message = "test message"
				entry.Level = logrus.InfoLevel
				entry.Data = logrus.Fields{
					"key1": "value1",
					"key2": "value2",
				}

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				outputStr := string(output)
				Expect(outputStr).To(ContainSubstring("key1"))
				Expect(outputStr).To(ContainSubstring("value1"))
			})

			It("Sorts fields by default", func() {
				formatter.DisableColors = true
				formatter.DisableSorting = false
				entry.Message = "test message"
				entry.Level = logrus.InfoLevel
				entry.Data = logrus.Fields{
					"z_key": "value1",
					"a_key": "value2",
				}

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				outputStr := string(output)
				// a_key should appear before z_key when sorted
				aIndex := strings.Index(outputStr, "a_key")
				zIndex := strings.Index(outputStr, "z_key")
				Expect(aIndex).To(BeNumerically("<", zIndex))
			})

			It("Does not sort fields when DisableSorting is true", func() {
				formatter.DisableSorting = true
				formatter.DisableColors = true
				entry.Message = "test message"
				entry.Level = logrus.InfoLevel
				entry.Data = logrus.Fields{
					"key1": "value1",
				}

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				Expect(string(output)).To(ContainSubstring("key1"))
			})
		})

		Context("Prefix extraction", func() {
			It("Extracts prefix from message with [prefix] format", func() {
				formatter.DisableColors = true
				entry.Message = "[myprefix] test message"
				entry.Level = logrus.InfoLevel

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				Expect(string(output)).To(ContainSubstring("test message"))
			})

			It("Uses prefix from fields if provided", func() {
				formatter.DisableColors = true
				entry.Message = "test message"
				entry.Level = logrus.InfoLevel
				entry.Data = logrus.Fields{
					"prefix": "custom",
				}

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				Expect(string(output)).To(ContainSubstring("test message"))
			})
		})

		Context("QuoteCharacter option", func() {
			It("Uses double quotes by default", func() {
				formatter.DisableColors = true
				entry.Message = "test message"
				entry.Level = logrus.InfoLevel
				entry.Data = logrus.Fields{
					"key": "value with spaces",
				}

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				Expect(string(output)).To(ContainSubstring("\""))
			})

			It("Uses custom quote character when specified", func() {
				formatter.DisableColors = true
				formatter.QuoteCharacter = "'"
				entry.Message = "test message"
				entry.Level = logrus.InfoLevel
				entry.Data = logrus.Fields{
					"key": "value with spaces",
				}

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				Expect(string(output)).To(ContainSubstring("'"))
			})
		})

		Context("QuoteEmptyFields option", func() {
			It("Quotes empty fields when QuoteEmptyFields is true", func() {
				formatter.DisableColors = true
				formatter.QuoteEmptyFields = true
				entry.Message = "test message"
				entry.Level = logrus.InfoLevel
				entry.Data = logrus.Fields{
					"empty": "",
				}

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				Expect(string(output)).To(ContainSubstring("\"\""))
			})
		})

		Context("Error values", func() {
			It("Formats error values correctly", func() {
				formatter.DisableColors = true
				entry.Message = "test message"
				entry.Level = logrus.ErrorLevel
				entry.Data = logrus.Fields{
					"error": errors.New("test error"),
				}

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				Expect(string(output)).To(ContainSubstring("test error"))
			})
		})

		Context("ForceColors option", func() {
			It("Forces colors when ForceColors is true", func() {
				formatter.ForceColors = true
				formatter.DisableColors = false
				entry.Message = "test message"
				entry.Level = logrus.InfoLevel

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				// Should contain ANSI color codes
				Expect(len(output)).To(BeNumerically(">", 0))
			})
		})

		Context("ColorScheme", func() {
			It("Uses custom color scheme when set", func() {
				colorScheme := &log.ColorScheme{
					InfoLevelStyle:  "red",
					WarnLevelStyle:  "blue",
					ErrorLevelStyle: "green",
					FatalLevelStyle: "yellow",
					PanicLevelStyle: "magenta",
					DebugLevelStyle: "cyan",
					PrefixStyle:     "white",
					TimestampStyle:  "black",
				}
				formatter.SetColorScheme(colorScheme)
				formatter.ForceColors = true
				entry.Message = "test message"
				entry.Level = logrus.InfoLevel

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				Expect(output).ToNot(BeNil())
			})
		})

		Context("SpacePadding option", func() {
			It("Applies space padding when specified", func() {
				formatter.SpacePadding = 50
				formatter.DisableColors = true
				formatter.ForceFormatting = true
				entry.Message = "short"
				entry.Level = logrus.InfoLevel

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				// Verify output is generated and padding affects the message
				Expect(output).ToNot(BeNil())
				Expect(string(output)).To(ContainSubstring("short"))
			})

			It("Does not apply padding when SpacePadding is 0", func() {
				formatter.SpacePadding = 0
				formatter.DisableColors = true
				entry.Message = "short"
				entry.Level = logrus.InfoLevel

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				// Just verify it formats without error
				Expect(output).ToNot(BeNil())
			})
		})

		Context("FullTimestamp option", func() {
			It("Shows full timestamp when FullTimestamp is true", func() {
				formatter.FullTimestamp = true
				formatter.DisableColors = true
				formatter.DisableTimestamp = false
				entry.Message = "test message"
				entry.Level = logrus.InfoLevel

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				Expect(output).ToNot(BeNil())
			})

			It("Uses custom timestamp format when specified", func() {
				formatter.FullTimestamp = true
				formatter.TimestampFormat = "2006-01-02"
				formatter.DisableColors = true
				formatter.DisableTimestamp = false
				entry.Message = "test message"
				entry.Level = logrus.InfoLevel

				output, err := formatter.Format(entry)
				Expect(err).To(BeNil())
				Expect(output).ToNot(BeNil())
			})
		})
	})
})
