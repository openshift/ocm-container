package pagerduty

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
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

		It("Returns an error when viper cannot unmarshal the config", func() {
			viper.Set("features.pagerduty", map[string]any{
				"enabled": "someString",
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("decoding failed"))
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

	Context("Tests Feature.Initialize()", func() {
		It("Returns OptionSet with volume mount when config file exists", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			configFile := "/path/to/my/config.json"
			err := afs.WriteFile(configFile, []byte("{}"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:   true,
					FilePath:  configFile,
					MountOpts: "rw",
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(1))
			Expect(opts.Mounts[0].Source).To(Equal(configFile))
			Expect(opts.Mounts[0].Destination).To(Equal(pagerDutyTokenDest))
			Expect(opts.Mounts[0].MountOptions).To(Equal("rw"))
		})

		It("Returns OptionSet with ro mount option when configured", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			configFile := "/path/to/my/config.json"
			err := afs.WriteFile(configFile, []byte("{}"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:   true,
					FilePath:  configFile,
					MountOpts: "ro",
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(1))
			Expect(opts.Mounts[0].MountOptions).To(Equal("ro"))
		})

		It("Returns error when config file does not exist", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:   true,
					FilePath:  "/nonexistent/path/config.json",
					MountOpts: "rw",
				},
			}

			opts, err := f.Initialize()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("could not find"))
			Expect(opts.Mounts).To(HaveLen(0))
		})

		It("Finds config file in HOME directory", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			homeDir := os.Getenv("HOME")
			configDir := homeDir + "/.config/pagerduty-cli"
			err := afs.MkdirAll(configDir, 0755)
			Expect(err).To(BeNil())

			configFile := configDir + "/config.json"
			err = afs.WriteFile(configFile, []byte("{}"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:   true,
					FilePath:  defaultPagerDutyTokenFile,
					MountOpts: "rw",
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(1))
			Expect(opts.Mounts[0].Source).To(Equal(configFile))
		})
	})
})
