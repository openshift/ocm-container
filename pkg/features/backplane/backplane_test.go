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
			Expect(cfg.ConfigFile).To(Equal(defaultBackplaneConfig))
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
			Expect(f.config.ConfigFile).To(Equal(defaultBackplaneConfig))
		})

		It("Overwrites defaults with viper config", func() {
			viper.Set("features.backplane", map[string]any{
				"enabled":     false,
				"config_file": "/custom/backplane/config.json",
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).To(BeNil())

			Expect(f.userHasConfig).To(BeTrue())
			Expect(f.config.Enabled).To(BeFalse())
			Expect(f.config.ConfigFile).To(Equal("/custom/backplane/config.json"))
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

	Context("Tests statFileLocation()", func() {
		It("Returns absolute path when it exists", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			configPath := "/absolute/backplane/config.json"

			err := afs.MkdirAll("/absolute/backplane", 0755)
			Expect(err).To(BeNil())
			err = afs.WriteFile(configPath, []byte("{}"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					ConfigFile: configPath,
				},
			}

			result, err := f.statFileLocation(configPath)
			Expect(err).To(BeNil())
			Expect(result).To(Equal(configPath))
		})

		It("Returns HOME-relative path when it exists", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			homeDir := os.Getenv("HOME")
			relativeFile := ".config/backplane/config.json"
			fullPath := homeDir + "/" + relativeFile

			err := afs.MkdirAll(homeDir+"/.config/backplane", 0755)
			Expect(err).To(BeNil())
			err = afs.WriteFile(fullPath, []byte("{}"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					ConfigFile: relativeFile,
				},
			}

			result, err := f.statFileLocation(relativeFile)
			Expect(err).To(BeNil())
			Expect(result).To(Equal(fullPath))
		})

		It("Returns error when file doesn't exist in any location", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			relativeFile := ".config/backplane/nonexistent.json"

			f := Feature{
				afs: &afs,
				config: &config{
					ConfigFile: relativeFile,
				},
			}

			result, err := f.statFileLocation(relativeFile)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("could not find"))
			Expect(result).To(Equal(""))
		})

		It("Returns error when filepath is empty", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}

			f := Feature{
				afs: &afs,
				config: &config{
					ConfigFile: "",
				},
			}

			result, err := f.statFileLocation("")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("no filepath provided"))
			Expect(result).To(Equal(""))
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
					Enabled:    true,
					ConfigFile: defaultBackplaneConfig,
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
					Enabled:    true,
					ConfigFile: defaultBackplaneConfig,
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

		It("Creates volume mount using custom config_file setting", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			customConfigFile := ".backplane/custom-config.json"
			homeDir := os.Getenv("HOME")
			fullPath := homeDir + "/" + customConfigFile

			// Ensure BACKPLANE_CONFIG is not set
			os.Unsetenv("BACKPLANE_CONFIG")

			// Create the custom config file
			err := afs.MkdirAll(homeDir+"/.backplane", 0755)
			Expect(err).To(BeNil())
			err = afs.WriteFile(fullPath, []byte("{}"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:    true,
					ConfigFile: customConfigFile,
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(1))
			Expect(opts.Mounts[0].Source).To(Equal(fullPath))
			Expect(opts.Mounts[0].Destination).To(Equal(backplaneConfigDest))
			Expect(opts.Mounts[0].MountOptions).To(Equal(backplaneConfigMountOpts))
			Expect(opts.Envs).To(HaveLen(1))
			Expect(opts.Envs[0].Key).To(Equal("BACKPLANE_CONFIG"))
			Expect(opts.Envs[0].Value).To(Equal(backplaneConfigDest))
		})

		It("BACKPLANE_CONFIG env var takes priority over config_file setting", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			envConfigPath := "/env/backplane/config.json"
			customConfigFile := ".backplane/custom-config.json"

			// Create both files
			err := afs.MkdirAll("/env/backplane", 0755)
			Expect(err).To(BeNil())
			err = afs.WriteFile(envConfigPath, []byte("{}"), 0644)
			Expect(err).To(BeNil())

			homeDir := os.Getenv("HOME")
			err = afs.MkdirAll(homeDir+"/.backplane", 0755)
			Expect(err).To(BeNil())
			err = afs.WriteFile(homeDir+"/"+customConfigFile, []byte("{}"), 0644)
			Expect(err).To(BeNil())

			// Set BACKPLANE_CONFIG environment variable
			os.Setenv("BACKPLANE_CONFIG", envConfigPath)
			defer os.Unsetenv("BACKPLANE_CONFIG")

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:    true,
					ConfigFile: customConfigFile,
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(1))
			// Should use env var path, not config_file
			Expect(opts.Mounts[0].Source).To(Equal(envConfigPath))
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
