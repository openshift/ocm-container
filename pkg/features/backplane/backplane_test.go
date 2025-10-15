package backplane

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

var _ = Describe("Pkg/Features/Backplane/Backplane", func() {
	BeforeEach(func() {
		viper.Reset()
	})

	Context("Tests the config", func() {
		It("Builds the defaults correctly", func() {
			cfg := newConfigWithDefaults()
			Expect(cfg).ToNot(BeNil())
			Expect(cfg.Enabled).To(BeTrue())
		})
	})

	Context("Tests config.validate()", func() {
		It("Returns nil for valid config", func() {
			cfg := config{
				Enabled: true,
			}
			err := cfg.validate()
			Expect(err).To(BeNil())
		})

		It("Returns nil for disabled config", func() {
			cfg := config{
				Enabled: false,
			}
			err := cfg.validate()
			Expect(err).To(BeNil())
		})
	})

	Context("Tests Feature.Configure()", func() {
		It("Uses defaults when no config is set", func() {
			f := Feature{}
			err := f.Configure()
			Expect(err).To(BeNil())

			Expect(f.userHasConfig).To(BeFalse())
			Expect(f.config.Enabled).To(BeTrue())
		})

		It("Overwrites defaults with viper config", func() {
			viper.Set("features.backplane", map[string]any{
				"enabled": false,
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).To(BeNil())

			Expect(f.userHasConfig).To(BeTrue())
			Expect(f.config.Enabled).To(BeFalse())
		})

		It("Returns an error when viper cannot unmarshal the config", func() {
			viper.Set("features.backplane", map[string]any{
				"enabled": "someString",
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("decoding failed"))
		})
	})

	Context("Tests Feature.Enabled()", func() {
		It("Returns true when config is enabled", func() {
			f := Feature{
				config: &config{Enabled: true},
			}
			Expect(f.Enabled()).To(BeTrue())
		})

		It("Returns false when config is disabled", func() {
			f := Feature{
				config: &config{Enabled: false},
			}
			Expect(f.Enabled()).To(BeFalse())
		})

		It("Returns false when feature flag is set", func() {
			viper.Set(FeatureFlagName, true)
			f := Feature{
				config: &config{Enabled: true},
			}
			Expect(f.Enabled()).To(BeFalse())
		})

		It("Returns false when both config is disabled and flag is set", func() {
			viper.Set(FeatureFlagName, true)
			f := Feature{
				config: &config{Enabled: false},
			}
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
		It("Creates volume mount and env var when BACKPLANE_CONFIG is set and file exists", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			configPath := "/custom/backplane/config.json"

			// Create the config file
			err := afs.MkdirAll("/custom/backplane", 0755)
			Expect(err).To(BeNil())
			err = afs.WriteFile(configPath, []byte("{}"), 0644)
			Expect(err).To(BeNil())

			// Set BACKPLANE_CONFIG environment variable
			os.Setenv("BACKPLANE_CONFIG", configPath)
			defer os.Unsetenv("BACKPLANE_CONFIG")

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled: true,
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(1))
			Expect(opts.Mounts[0].Source).To(Equal(configPath))
			Expect(opts.Mounts[0].Destination).To(Equal(backplaneConfigDest))
			Expect(opts.Mounts[0].MountOptions).To(Equal(backplaneConfigMountOpts))
			Expect(opts.Envs).To(HaveLen(1))
			Expect(opts.Envs[0].Key).To(Equal("BACKPLANE_CONFIG"))
			Expect(opts.Envs[0].Value).To(Equal(backplaneConfigDest))
		})

		It("Creates volume mount and env var when using default config location", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			homeDir := os.Getenv("HOME")
			defaultPath := homeDir + "/" + defaultBackplaneConfig

			// Create the default config file
			err := afs.MkdirAll(homeDir+"/.config/backplane", 0755)
			Expect(err).To(BeNil())
			err = afs.WriteFile(defaultPath, []byte("{}"), 0644)
			Expect(err).To(BeNil())

			// Ensure BACKPLANE_CONFIG is not set
			os.Unsetenv("BACKPLANE_CONFIG")

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled: true,
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(1))
			Expect(opts.Mounts[0].Source).To(Equal(defaultPath))
			Expect(opts.Mounts[0].Destination).To(Equal(backplaneConfigDest))
			Expect(opts.Mounts[0].MountOptions).To(Equal(backplaneConfigMountOpts))
			Expect(opts.Envs).To(HaveLen(1))
			Expect(opts.Envs[0].Key).To(Equal("BACKPLANE_CONFIG"))
			Expect(opts.Envs[0].Value).To(Equal(backplaneConfigDest))
		})

		It("Returns error when config file doesn't exist", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			homeDir := os.Getenv("HOME")
			defaultPath := homeDir + "/" + defaultBackplaneConfig

			// Ensure BACKPLANE_CONFIG is not set and file doesn't exist
			os.Unsetenv("BACKPLANE_CONFIG")

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled: true,
				},
			}

			// Verify file doesn't exist
			_, err := afs.Stat(defaultPath)
			Expect(err).ToNot(BeNil())

			opts, err := f.Initialize()
			Expect(err).ToNot(BeNil())
			Expect(opts.Mounts).To(HaveLen(0))
			Expect(opts.Envs).To(HaveLen(0))
		})

		It("Returns error when BACKPLANE_CONFIG file doesn't exist", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			configPath := "/nonexistent/backplane/config.json"

			// Set BACKPLANE_CONFIG to nonexistent file
			os.Setenv("BACKPLANE_CONFIG", configPath)
			defer os.Unsetenv("BACKPLANE_CONFIG")

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled: true,
				},
			}

			// Verify file doesn't exist
			_, err := afs.Stat(configPath)
			Expect(err).ToNot(BeNil())

			opts, err := f.Initialize()
			Expect(err).ToNot(BeNil())
			Expect(opts.Mounts).To(HaveLen(0))
			Expect(opts.Envs).To(HaveLen(0))
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
