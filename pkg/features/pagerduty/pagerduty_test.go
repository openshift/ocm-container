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
			Expect(f.config.MountOpts).To(Equal("ro"))
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
		DescribeTable("Valid mount options",
			func(mountOpt string) {
				cfg := config{MountOpts: mountOpt}
				err := cfg.validate()
				Expect(err).To(BeNil())
			},
			Entry("ro", "ro"),
			Entry("rw", "rw"),
			Entry("z", "z"),
			Entry("Z", "Z"),
			Entry("ro,z", "ro,z"),
			Entry("ro,Z", "ro,Z"),
			Entry("rw,z", "rw,z"),
			Entry("rw,Z", "rw,Z"),
		)

		DescribeTable("Invalid mount options",
			func(mountOpt string) {
				cfg := config{MountOpts: mountOpt}
				err := cfg.validate()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("invalid mount option"))
			},
			Entry("invalid option", "invalid"),
			Entry("empty option", ""),
		)
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

		DescribeTable("Mounts with various mount options",
			func(mountOpt string) {
				afs := afero.Afero{Fs: afero.NewMemMapFs()}
				configFile := "/test/config.json"

				err := afs.WriteFile(configFile, []byte("{}"), 0644)
				Expect(err).To(BeNil())

				f := Feature{
					afs: &afs,
					config: &config{
						Enabled:   true,
						FilePath:  configFile,
						MountOpts: mountOpt,
					},
				}

				opts, err := f.Initialize()
				Expect(err).To(BeNil())
				Expect(opts.Mounts[0].MountOptions).To(Equal(mountOpt))
			},
			Entry("ro", "ro"),
			Entry("rw", "rw"),
			Entry("rw,z", "rw,z"),
			Entry("ro,Z", "ro,Z"),
		)

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
			configDir := homeDir + "/.config/pagerduty"
			err := afs.MkdirAll(configDir, 0755)
			Expect(err).To(BeNil())

			configFile := configDir + "/token.json"
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

	Context("Tests statConfigFileLocations()", func() {
		It("Returns absolute path when it exists", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			configFile := "/absolute/path/config.json"
			err := afs.WriteFile(configFile, []byte("{}"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs:    &afs,
				config: &config{FilePath: configFile},
			}

			path, err := f.statConfigFileLocations()
			Expect(err).To(BeNil())
			Expect(path).To(Equal(configFile))
		})

		It("Returns HOME-relative path when absolute doesn't exist", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			homeDir := os.Getenv("HOME")
			relativePath := ".config/pagerduty/token.json"
			fullPath := homeDir + "/" + relativePath

			err := afs.MkdirAll(homeDir+"/.config/pagerduty", 0755)
			Expect(err).To(BeNil())
			err = afs.WriteFile(fullPath, []byte("{}"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs:    &afs,
				config: &config{FilePath: relativePath},
			}

			path, err := f.statConfigFileLocations()
			Expect(err).To(BeNil())
			Expect(path).To(Equal(fullPath))
		})

		It("Returns error when file not found in any location", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			f := Feature{
				afs:    &afs,
				config: &config{FilePath: "nonexistent.json"},
			}

			path, err := f.statConfigFileLocations()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("could not find"))
			Expect(path).To(Equal(""))
		})
	})

	Context("Tests Feature.HandleError()", func() {
		It("Does not panic when userHasConfig is true", func() {
			f := Feature{userHasConfig: true}
			Expect(func() {
				f.HandleError(os.ErrNotExist)
			}).NotTo(Panic())
		})

		It("Does not panic when userHasConfig is false", func() {
			f := Feature{userHasConfig: false}
			Expect(func() {
				f.HandleError(os.ErrNotExist)
			}).NotTo(Panic())
		})
	})
})
