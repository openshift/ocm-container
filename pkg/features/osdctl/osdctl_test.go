package osdctl

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

var _ = Describe("Pkg/Features/Osdctl/Osdctl", func() {
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
			Expect(f.config.ConfigFile).To(Equal(osdctlConfigFile))
			Expect(f.config.TokenFile).To(Equal(vaultTokenFile))
			Expect(f.config.ConfigMountOptions).To(Equal("ro"))
			Expect(f.config.TokenMountOptions).To(Equal("rw"))
		})

		It("Overwrites defaults with viper config", func() {
			viper.Set("features.osdctl", map[string]any{
				"enabled":              false,
				"config_file":          "/custom/osdctl/config",
				"token_file":           "/custom/token",
				"config_mount_options": "rw",
				"token_mount_options":  "ro",
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).To(BeNil())

			Expect(f.userHasConfig).To(BeTrue())
			Expect(f.config.Enabled).To(BeFalse())
			Expect(f.config.ConfigFile).To(Equal("/custom/osdctl/config"))
			Expect(f.config.TokenFile).To(Equal("/custom/token"))
			Expect(f.config.ConfigMountOptions).To(Equal("rw"))
			Expect(f.config.TokenMountOptions).To(Equal("ro"))
		})

		It("Returns an error when viper cannot unmarshal the config", func() {
			viper.Set("features.osdctl", map[string]any{
				"enabled": "someString",
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("decoding failed"))
		})
	})

	Context("Tests config.validate()", func() {
		It("Returns nil for valid config with ro config mount options", func() {
			cfg := config{
				Enabled:            true,
				ConfigFile:         ".config/osdctl",
				ConfigMountOptions: "ro",
				TokenMountOptions:  "rw",
			}
			err := cfg.validate()
			Expect(err).To(BeNil())
		})

		It("Returns nil for valid config with rw config mount options", func() {
			cfg := config{
				Enabled:            true,
				ConfigFile:         ".config/osdctl",
				ConfigMountOptions: "rw",
				TokenMountOptions:  "ro",
			}
			err := cfg.validate()
			Expect(err).To(BeNil())
		})

		It("Returns error for invalid config mount options", func() {
			cfg := config{
				Enabled:            true,
				ConfigFile:         ".config/osdctl",
				ConfigMountOptions: "invalid",
				TokenMountOptions:  "rw",
			}
			err := cfg.validate()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("config_mount_options must be either 'ro' or 'rw'"))
		})

		It("Returns error for invalid token mount options", func() {
			cfg := config{
				Enabled:            true,
				ConfigFile:         ".config/osdctl",
				ConfigMountOptions: "ro",
				TokenMountOptions:  "invalid",
			}
			err := cfg.validate()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("token_mount_options must be either 'ro' or 'rw'"))
		})

		It("Returns nil for disabled config", func() {
			cfg := config{
				Enabled:            false,
				ConfigMountOptions: "ro",
				TokenMountOptions:  "rw",
			}
			err := cfg.validate()
			Expect(err).To(BeNil())
		})
	})

	Context("Tests Feature.Enabled()", func() {
		It("Returns true when config is enabled, has user config, and config file is set", func() {
			f := Feature{
				config:        &config{Enabled: true, ConfigFile: osdctlConfigFile},
				userHasConfig: true,
			}
			Expect(f.Enabled()).To(BeTrue())
		})

		It("Returns false when config is disabled", func() {
			f := Feature{
				config:        &config{Enabled: false, ConfigFile: osdctlConfigFile},
				userHasConfig: true,
			}
			Expect(f.Enabled()).To(BeFalse())
		})

		It("Returns false when feature flag is set", func() {
			viper.Set(FeatureFlagName, true)
			f := Feature{
				config:        &config{Enabled: true, ConfigFile: osdctlConfigFile},
				userHasConfig: true,
			}
			Expect(f.Enabled()).To(BeFalse())
		})

		It("Returns false when user has no config", func() {
			f := Feature{
				config:        &config{Enabled: true, ConfigFile: osdctlConfigFile},
				userHasConfig: false,
			}
			Expect(f.Enabled()).To(BeFalse())
		})

		It("Returns false when config file is empty", func() {
			f := Feature{
				config:        &config{Enabled: true, ConfigFile: ""},
				userHasConfig: true,
			}
			Expect(f.Enabled()).To(BeFalse())
		})

		It("Returns false when both config is disabled and flag is set", func() {
			viper.Set(FeatureFlagName, true)
			f := Feature{
				config:        &config{Enabled: false, ConfigFile: osdctlConfigFile},
				userHasConfig: true,
			}
			Expect(f.Enabled()).To(BeFalse())
		})
	})

	Context("Tests Feature.ExitOnError()", func() {
		It("Returns false when criticalError is false", func() {
			f := Feature{criticalError: false}
			Expect(f.ExitOnError()).To(BeFalse())
		})

		It("Returns true when criticalError is true", func() {
			f := Feature{criticalError: true}
			Expect(f.ExitOnError()).To(BeTrue())
		})
	})

	Context("Tests Feature.Initialize()", func() {
		It("Returns OptionSet with volume mounts when both config and token files exist", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			configPath := "/absolute/path/.config/osdctl"
			tokenPath := "/absolute/path/.vault-token"

			err := afs.WriteFile(configPath, []byte("config"), 0644)
			Expect(err).To(BeNil())
			err = afs.WriteFile(tokenPath, []byte("token"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:            true,
					ConfigFile:         configPath,
					TokenFile:          tokenPath,
					ConfigMountOptions: "ro",
					TokenMountOptions:  "rw",
				},
				userHasConfig: true,
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(2))
			Expect(opts.Mounts[0].Source).To(Equal(configPath))
			Expect(opts.Mounts[0].Destination).To(Equal("/root/.config/osdctl"))
			Expect(opts.Mounts[0].MountOptions).To(Equal("ro"))
			Expect(opts.Mounts[1].Source).To(Equal(tokenPath))
			Expect(opts.Mounts[1].Destination).To(Equal("/root/.vault-token"))
			Expect(opts.Mounts[1].MountOptions).To(Equal("rw"))
		})

		It("Returns OptionSet with only config mount when token file does not exist", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			configPath := "/absolute/path/.config/osdctl"

			err := afs.WriteFile(configPath, []byte("config"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:            true,
					ConfigFile:         configPath,
					TokenFile:          "/nonexistent/.vault-token",
					ConfigMountOptions: "ro",
					TokenMountOptions:  "rw",
				},
				userHasConfig: true,
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(1))
			Expect(opts.Mounts[0].Source).To(Equal(configPath))
			Expect(opts.Mounts[0].Destination).To(Equal("/root/.config/osdctl"))
		})

		It("Finds config file in HOME directory", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			homeDir := os.Getenv("HOME")
			configPath := homeDir + "/.config/osdctl"

			err := afs.WriteFile(configPath, []byte("config"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:            true,
					ConfigFile:         osdctlConfigFile,
					TokenFile:          vaultTokenFile,
					ConfigMountOptions: "ro",
					TokenMountOptions:  "rw",
				},
				userHasConfig: true,
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(1))
			Expect(opts.Mounts[0].Source).To(Equal(configPath))
		})

		It("Returns error when config file does not exist", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:            true,
					ConfigFile:         osdctlConfigFile,
					TokenFile:          vaultTokenFile,
					ConfigMountOptions: "ro",
					TokenMountOptions:  "rw",
				},
				userHasConfig: false,
			}

			opts, err := f.Initialize()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("could not find"))
			Expect(opts.Mounts).To(HaveLen(0))
			Expect(f.criticalError).To(BeFalse())
		})

		It("Sets criticalError when user has config and config file does not exist", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:            true,
					ConfigFile:         osdctlConfigFile,
					TokenFile:          vaultTokenFile,
					ConfigMountOptions: "ro",
					TokenMountOptions:  "rw",
				},
				userHasConfig: true,
			}

			opts, err := f.Initialize()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("could not find"))
			Expect(opts.Mounts).To(HaveLen(0))
			Expect(f.criticalError).To(BeTrue())
		})

		It("Uses custom mount options", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			configPath := "/test/.config/osdctl"
			tokenPath := "/test/.vault-token"

			err := afs.WriteFile(configPath, []byte("config"), 0644)
			Expect(err).To(BeNil())
			err = afs.WriteFile(tokenPath, []byte("token"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:            true,
					ConfigFile:         configPath,
					TokenFile:          tokenPath,
					ConfigMountOptions: "rw",
					TokenMountOptions:  "ro",
				},
				userHasConfig: true,
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts[0].MountOptions).To(Equal("rw"))
			Expect(opts.Mounts[1].MountOptions).To(Equal("ro"))
		})

		It("Uses custom config file path when specified", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			customPath := "/custom/osdctl/config"

			err := afs.WriteFile(customPath, []byte("config"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:            true,
					ConfigFile:         customPath,
					TokenFile:          vaultTokenFile,
					ConfigMountOptions: "ro",
					TokenMountOptions:  "rw",
				},
				userHasConfig: true,
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts[0].Source).To(Equal(customPath))
		})
	})

	Context("Tests statFileLocations()", func() {
		It("Returns absolute path when it exists", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			absolutePath := "/absolute/path/.config/osdctl"

			err := afs.WriteFile(absolutePath, []byte("config"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
			}

			path, err := f.statFileLocations(absolutePath)
			Expect(err).To(BeNil())
			Expect(path).To(Equal(absolutePath))
		})

		It("Returns HOME-relative path when absolute doesn't exist", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			homeDir := os.Getenv("HOME")
			relativePath := ".config/osdctl"
			fullPath := homeDir + "/" + relativePath

			err := afs.WriteFile(fullPath, []byte("config"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
			}

			path, err := f.statFileLocations(relativePath)
			Expect(err).To(BeNil())
			Expect(path).To(Equal(fullPath))
		})

		It("Returns error when file not found in any location", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}

			f := Feature{
				afs: &afs,
			}

			path, err := f.statFileLocations("nonexistent")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("could not find"))
			Expect(path).To(Equal(""))
		})

		It("Returns error when filepath is empty", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}

			f := Feature{
				afs: &afs,
			}

			path, err := f.statFileLocations("")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("no filepath provided"))
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
