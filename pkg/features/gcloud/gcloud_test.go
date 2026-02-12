package gcloud

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

var _ = Describe("Pkg/Features/Gcloud/Gcloud", func() {
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
			Expect(f.config.ConfigDir).To(Equal(gcloudConfigDir))
		})

		It("Overwrites defaults with viper config", func() {
			viper.Set("features.gcloud", map[string]any{
				"enabled":    false,
				"config_dir": "/custom/gcloud/config",
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).To(BeNil())

			Expect(f.userHasConfig).To(BeTrue())
			Expect(f.config.Enabled).To(BeFalse())
			Expect(f.config.ConfigDir).To(Equal("/custom/gcloud/config"))
		})

		It("Returns an error when viper cannot unmarshal the config", func() {
			viper.Set("features.gcloud", map[string]any{
				"enabled": "someString",
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("decoding failed"))
		})

		It("Returns an error when validation fails due to invalid mount option", func() {
			viper.Set("features.gcloud", map[string]any{
				"enabled":      true,
				"config_dir":   "/custom/gcloud/config",
				"config_mount": "invalid-option",
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("invalid mount option"))
		})
	})

	Context("Tests config.validate()", func() {
		DescribeTable("Valid mount options",
			func(mountOpt string) {
				cfg := config{Enabled: true, ConfigDir: ".config/gcloud", MountOpts: mountOpt}
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
				cfg := config{Enabled: true, ConfigDir: ".config/gcloud", MountOpts: mountOpt}
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
		It("Returns OptionSet with volume mount when gcloud config dir exists", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			gcloudPath := "/absolute/path/.config/gcloud"

			err := afs.MkdirAll(gcloudPath, 0755)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:   true,
					ConfigDir: gcloudPath,
					MountOpts: "ro",
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(1))
			Expect(opts.Mounts[0].Source).To(Equal(gcloudPath))
			Expect(opts.Mounts[0].Destination).To(Equal("/root/.config/gcloud"))
			Expect(opts.Mounts[0].MountOptions).To(Equal("ro"))
		})

		It("Finds config dir in HOME directory", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			homeDir := os.Getenv("HOME")
			gcloudPath := homeDir + "/.config/gcloud"

			err := afs.MkdirAll(gcloudPath, 0755)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:   true,
					ConfigDir: gcloudConfigDir,
					MountOpts: "ro",
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(1))
			Expect(opts.Mounts[0].Source).To(Equal(gcloudPath))
		})

		It("Returns error when gcloud config dir does not exist", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:   true,
					ConfigDir: gcloudConfigDir,
					MountOpts: "ro",
				},
			}

			opts, err := f.Initialize()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("could not find"))
			Expect(opts.Mounts).To(HaveLen(0))
		})

		It("Returns error when config dir is empty", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:   true,
					ConfigDir: "",
					MountOpts: "ro",
				},
			}

			opts, err := f.Initialize()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("no filepath provided"))
			Expect(opts.Mounts).To(HaveLen(0))
		})

		It("Uses custom config dir when specified", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			customPath := "/custom/gcloud/config"

			err := afs.MkdirAll(customPath, 0755)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:   true,
					ConfigDir: customPath,
					MountOpts: "ro",
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts[0].Source).To(Equal(customPath))
		})

		DescribeTable("Mounts with various mount options",
			func(mountOpt string) {
				afs := afero.Afero{Fs: afero.NewMemMapFs()}
				gcloudPath := "/test/.config/gcloud"

				err := afs.MkdirAll(gcloudPath, 0755)
				Expect(err).To(BeNil())

				f := Feature{
					afs: &afs,
					config: &config{
						Enabled:   true,
						ConfigDir: gcloudPath,
						MountOpts: mountOpt,
					},
				}

				opts, err := f.Initialize()
				Expect(err).To(BeNil())
				Expect(opts.Mounts[0].MountOptions).To(Equal(mountOpt))
			},
			Entry("ro", "ro"),
			Entry("rw,z", "rw,z"),
			Entry("ro,Z", "ro,Z"),
		)
	})

	Context("Tests statConfigFileLocations()", func() {
		It("Returns absolute path when it exists", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			absolutePath := "/absolute/path/.config/gcloud"

			err := afs.MkdirAll(absolutePath, 0755)
			Expect(err).To(BeNil())

			f := Feature{
				afs:    &afs,
				config: &config{ConfigDir: absolutePath},
			}

			path, err := f.statConfigFileLocations()
			Expect(err).To(BeNil())
			Expect(path).To(Equal(absolutePath))
		})

		It("Returns HOME-relative path when absolute doesn't exist", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			homeDir := os.Getenv("HOME")
			relativePath := ".config/gcloud"
			fullPath := homeDir + "/" + relativePath

			err := afs.MkdirAll(fullPath, 0755)
			Expect(err).To(BeNil())

			f := Feature{
				afs:    &afs,
				config: &config{ConfigDir: relativePath},
			}

			path, err := f.statConfigFileLocations()
			Expect(err).To(BeNil())
			Expect(path).To(Equal(fullPath))
		})

		It("Returns error when config dir not found in any location", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}

			f := Feature{
				afs:    &afs,
				config: &config{ConfigDir: "nonexistent"},
			}

			path, err := f.statConfigFileLocations()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("could not find"))
			Expect(path).To(Equal(""))
		})

		It("Returns error when config dir is empty", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}

			f := Feature{
				afs:    &afs,
				config: &config{ConfigDir: ""},
			}

			path, err := f.statConfigFileLocations()
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
