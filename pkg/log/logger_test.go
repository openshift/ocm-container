package log_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/ocm-container/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var _ = Describe("Pkg/Log/Logger", func() {
	BeforeEach(func() {
		viper.Reset()
		// Reset logrus to default state
		logrus.SetLevel(logrus.InfoLevel)
		logrus.SetFormatter(&logrus.TextFormatter{})
	})

	Describe("InitializeLogger", func() {
		Context("Log level configuration", func() {
			It("Sets default WarnLevel when no config is provided", func() {
				err := log.InitializeLogger()
				Expect(err).To(BeNil())
				Expect(logrus.GetLevel()).To(Equal(logrus.WarnLevel))
			})

			It("Sets log level from 'log-level' flag", func() {
				viper.Set("log-level", "debug")
				err := log.InitializeLogger()
				Expect(err).To(BeNil())
				Expect(logrus.GetLevel()).To(Equal(logrus.DebugLevel))
			})

			It("Sets log level from 'log.level' config", func() {
				viper.Set("log.level", "info")
				err := log.InitializeLogger()
				Expect(err).To(BeNil())
				Expect(logrus.GetLevel()).To(Equal(logrus.InfoLevel))
			})

			It("Prioritizes 'log-level' flag over 'log.level' config", func() {
				viper.Set("log-level", "debug")
				viper.Set("log.level", "error")
				err := log.InitializeLogger()
				Expect(err).To(BeNil())
				Expect(logrus.GetLevel()).To(Equal(logrus.DebugLevel))
			})

			It("Returns error for invalid log level from flag", func() {
				viper.Set("log-level", "invalid")
				err := log.InitializeLogger()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("Invalid log level"))
			})

			It("Returns error for invalid log level from config", func() {
				viper.Set("log.level", "invalid")
				err := log.InitializeLogger()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("Invalid log level"))
			})
		})

		Context("Color configuration", func() {
			It("Disables colors when 'no-color' flag is set", func() {
				viper.Set("no-color", true)
				err := log.InitializeLogger()
				Expect(err).To(BeNil())

				formatter, ok := logrus.StandardLogger().Formatter.(*log.TextFormatter)
				Expect(ok).To(BeTrue())
				Expect(formatter.DisableColors).To(BeTrue())
			})

			It("Disables colors when 'log.color' is false", func() {
				viper.Set("log.color", false)
				err := log.InitializeLogger()
				Expect(err).To(BeNil())

				formatter, ok := logrus.StandardLogger().Formatter.(*log.TextFormatter)
				Expect(ok).To(BeTrue())
				Expect(formatter.DisableColors).To(BeTrue())
			})

			It("Enables colors when 'log.color' is true and 'no-color' is not set", func() {
				viper.Set("log.color", true)
				err := log.InitializeLogger()
				Expect(err).To(BeNil())

				formatter, ok := logrus.StandardLogger().Formatter.(*log.TextFormatter)
				Expect(ok).To(BeTrue())
				Expect(formatter.DisableColors).To(BeFalse())
			})

			It("Disables colors when both 'no-color' is true and 'log.color' is true (no-color takes precedence)", func() {
				viper.Set("no-color", true)
				viper.Set("log.color", true)
				err := log.InitializeLogger()
				Expect(err).To(BeNil())

				formatter, ok := logrus.StandardLogger().Formatter.(*log.TextFormatter)
				Expect(ok).To(BeTrue())
				Expect(formatter.DisableColors).To(BeTrue())
			})

			It("Enables colors by default when no color config is set", func() {
				err := log.InitializeLogger()
				Expect(err).To(BeNil())

				formatter, ok := logrus.StandardLogger().Formatter.(*log.TextFormatter)
				Expect(ok).To(BeTrue())
				Expect(formatter.DisableColors).To(BeFalse())
			})
		})

		Context("Formatter configuration", func() {
			It("Sets TextFormatter", func() {
				err := log.InitializeLogger()
				Expect(err).To(BeNil())

				_, ok := logrus.StandardLogger().Formatter.(*log.TextFormatter)
				Expect(ok).To(BeTrue())
			})

			It("Disables timestamp in formatter", func() {
				err := log.InitializeLogger()
				Expect(err).To(BeNil())

				formatter, ok := logrus.StandardLogger().Formatter.(*log.TextFormatter)
				Expect(ok).To(BeTrue())
				Expect(formatter.DisableTimestamp).To(BeTrue())
			})
		})

		Context("Combined configurations", func() {
			It("Sets debug level and no colors", func() {
				viper.Set("log-level", "debug")
				viper.Set("no-color", true)
				err := log.InitializeLogger()
				Expect(err).To(BeNil())

				Expect(logrus.GetLevel()).To(Equal(logrus.DebugLevel))
				formatter, ok := logrus.StandardLogger().Formatter.(*log.TextFormatter)
				Expect(ok).To(BeTrue())
				Expect(formatter.DisableColors).To(BeTrue())
			})

			It("Sets error level with colors enabled", func() {
				viper.Set("log.level", "error")
				viper.Set("log.color", true)
				err := log.InitializeLogger()
				Expect(err).To(BeNil())

				Expect(logrus.GetLevel()).To(Equal(logrus.ErrorLevel))
				formatter, ok := logrus.StandardLogger().Formatter.(*log.TextFormatter)
				Expect(ok).To(BeTrue())
				Expect(formatter.DisableColors).To(BeFalse())
			})
		})
	})
})
