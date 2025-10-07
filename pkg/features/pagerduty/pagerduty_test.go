package pagerduty

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

var _ = Describe("Pkg/Features/Pagerduty/Pagerduty", func() {
	BeforeEach(func() {
		viper.Reset()
	})

	Context("Tests the config", func() {
		It("Builds the defaults correctly", func() {
			f := Feature{}
			err := f.Configure()
			Expect(err).To(BeNil())

			Expect(f.userHasConfig).To(BeFalse())
			Expect(f.config.Enabled).To(BeTrue())
			Expect(f.config.FilePath).To(Equal(defaultPagerDutyTokenFile))
			Expect(f.config.MountOpts).To(Equal("rw"))
		})

		It("Overwrites defaults with viper config", func() {
			viper.Set("features.pagerduty", map[string]any{
				"enabled":      false,
				"config_file":  "/path/to/file",
				"config_mount": "ro",
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).To(BeNil())

			Expect(f.userHasConfig).To(BeTrue())
			Expect(f.config.Enabled).To(BeFalse())
			Expect(f.config.FilePath).To(Equal("/path/to/file"))
			Expect(f.config.MountOpts).To(Equal("ro"))
		})

		It("Returns error on invalid mount options", func() {
			viper.Set("features.pagerduty", map[string]any{
				"config_mount": "invalid",
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("invalid mount option"))
		})
	})

	Context("Tests config.validate()", func() {
		It("Accepts valid mount options 'ro'", func() {
			cfg := config{MountOpts: "ro"}
			err := cfg.validate()
			Expect(err).To(BeNil())
		})

		It("Accepts valid mount options 'rw'", func() {
			cfg := config{MountOpts: "rw"}
			err := cfg.validate()
			Expect(err).To(BeNil())
		})

		It("Rejects invalid mount options", func() {
			cfg := config{MountOpts: "invalid"}
			err := cfg.validate()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("invalid mount option"))
		})
	})

	Context("Tests Feature.Enabled()", func() {
		It("Returns true when config is enabled and flag is not set", func() {
			f := Feature{config: &config{Enabled: true}}
			Expect(f.Enabled()).To(BeTrue())
		})

		It("Returns false when config is disabled", func() {
			f := Feature{config: &config{Enabled: false}}
			Expect(f.Enabled()).To(BeFalse())
		})

		It("Returns false when feature flag is set", func() {
			viper.Set(FeatureFlagName, true)
			f := Feature{config: &config{Enabled: true}}
			Expect(f.Enabled()).To(BeFalse())
		})

		It("Returns false when both config is disabled and flag is set", func() {
			viper.Set(FeatureFlagName, true)
			f := Feature{config: &config{Enabled: false}}
			Expect(f.Enabled()).To(BeFalse())
		})
	})

	Context("Tests Feature.ExitOnError()", func() {
		It("Returns false", func() {
			f := Feature{}
			Expect(f.ExitOnError()).To(BeFalse())
		})
	})
})
