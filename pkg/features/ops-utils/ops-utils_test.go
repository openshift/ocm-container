package opsutils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

var _ = Describe("Pkg/Features/OpsUtils/OpsUtils", func() {
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
			Expect(f.config.MountOptions).To(Equal(defaultMountOptions))
			Expect(f.config.SourceDir).To(Equal(""))
		})

		It("Overwrites defaults with viper config", func() {
			viper.Set("features.ops_utils", map[string]any{
				"enabled":       false,
				"source_dir":    "/custom/ops/utils",
				"mount_options": "rw",
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).To(BeNil())

			Expect(f.userHasConfig).To(BeTrue())
			Expect(f.config.Enabled).To(BeFalse())
			Expect(f.config.SourceDir).To(Equal("/custom/ops/utils"))
			Expect(f.config.MountOptions).To(Equal("rw"))
		})

		It("Returns an error when viper cannot unmarshal the config", func() {
			viper.Set("features.ops_utils", map[string]any{
				"enabled": "someString",
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("decoding failed"))
		})
	})

	Context("Tests config.validate()", func() {
		It("Returns nil for valid 'ro' mount option", func() {
			cfg := config{Enabled: true, SourceDir: "/some/path", MountOptions: "ro"}
			err := cfg.validate()
			Expect(err).To(BeNil())
		})

		It("Returns nil for valid 'rw' mount option", func() {
			cfg := config{Enabled: true, SourceDir: "/some/path", MountOptions: "rw"}
			err := cfg.validate()
			Expect(err).To(BeNil())
		})

		It("Returns error for invalid mount option", func() {
			cfg := config{Enabled: true, SourceDir: "/some/path", MountOptions: "invalid"}
			err := cfg.validate()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("mount_options must be either 'ro' or 'rw'"))
		})

		It("Returns error for empty mount option", func() {
			cfg := config{Enabled: true, SourceDir: "/some/path", MountOptions: ""}
			err := cfg.validate()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("mount_options must be either 'ro' or 'rw'"))
		})
	})

	Context("Tests Feature.Enabled()", func() {
		It("Returns true when all conditions are met", func() {
			f := Feature{
				config:        &config{Enabled: true, SourceDir: "/some/path"},
				userHasConfig: true,
			}
			Expect(f.Enabled()).To(BeTrue())
		})

		It("Returns false when config is disabled", func() {
			f := Feature{
				config:        &config{Enabled: false, SourceDir: "/some/path"},
				userHasConfig: true,
			}
			Expect(f.Enabled()).To(BeFalse())
		})

		It("Returns false when feature flag is set", func() {
			viper.Set(FeatureFlagName, true)
			f := Feature{
				config:        &config{Enabled: true, SourceDir: "/some/path"},
				userHasConfig: true,
			}
			Expect(f.Enabled()).To(BeFalse())
		})

		It("Returns false when userHasConfig is false", func() {
			f := Feature{
				config:        &config{Enabled: true, SourceDir: "/some/path"},
				userHasConfig: false,
			}
			Expect(f.Enabled()).To(BeFalse())
		})

		It("Returns false when source_dir is empty", func() {
			f := Feature{
				config:        &config{Enabled: true, SourceDir: ""},
				userHasConfig: true,
			}
			Expect(f.Enabled()).To(BeFalse())
		})

		It("Returns false when config is disabled and flag is set", func() {
			viper.Set(FeatureFlagName, true)
			f := Feature{
				config:        &config{Enabled: false, SourceDir: "/some/path"},
				userHasConfig: true,
			}
			Expect(f.Enabled()).To(BeFalse())
		})
	})

	Context("Tests Feature.ExitOnError()", func() {
		It("Returns true", func() {
			f := Feature{}
			Expect(f.ExitOnError()).To(BeTrue())
		})
	})

	Context("Tests Feature.Initialize()", func() {
		It("Returns OptionSet with volume mount when source dir exists", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			opsUtilsPath := "/path/to/ops/utils"

			err := afs.MkdirAll(opsUtilsPath, 0755)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:      true,
					SourceDir:    opsUtilsPath,
					MountOptions: "ro",
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(1))
			Expect(opts.Mounts[0].Source).To(Equal(opsUtilsPath))
			Expect(opts.Mounts[0].Destination).To(Equal(destinationDir))
			Expect(opts.Mounts[0].MountOptions).To(Equal("ro"))
		})

		It("Returns error when source dir does not exist", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:      true,
					SourceDir:    "/nonexistent/path",
					MountOptions: "ro",
				},
			}

			opts, err := f.Initialize()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("problem reading Ops Utils directory"))
			Expect(opts.Mounts).To(HaveLen(0))
		})

		It("Mounts with ro option by default", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			opsUtilsPath := "/test/ops/utils"

			err := afs.MkdirAll(opsUtilsPath, 0755)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:      true,
					SourceDir:    opsUtilsPath,
					MountOptions: "ro",
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts[0].MountOptions).To(Equal("ro"))
		})

		It("Mounts with rw option when configured", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			opsUtilsPath := "/test/ops/utils"

			err := afs.MkdirAll(opsUtilsPath, 0755)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:      true,
					SourceDir:    opsUtilsPath,
					MountOptions: "rw",
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts[0].MountOptions).To(Equal("rw"))
		})

		It("Uses custom source dir when specified", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			customPath := "/custom/ops-utils/dir"

			err := afs.MkdirAll(customPath, 0755)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:      true,
					SourceDir:    customPath,
					MountOptions: "ro",
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts[0].Source).To(Equal(customPath))
		})

		It("Always mounts to /root/ops-utils destination", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			opsUtilsPath := "/any/path"

			err := afs.MkdirAll(opsUtilsPath, 0755)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:      true,
					SourceDir:    opsUtilsPath,
					MountOptions: "ro",
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts[0].Destination).To(Equal("/root/ops-utils"))
		})
	})

	Context("Tests Feature.HandleError()", func() {
		It("Does not panic with userHasConfig=true", func() {
			f := Feature{userHasConfig: true}
			Expect(func() {
				f.HandleError(nil)
			}).ToNot(Panic())
		})

		It("Does not panic with userHasConfig=false", func() {
			f := Feature{userHasConfig: false}
			Expect(func() {
				f.HandleError(nil)
			}).ToNot(Panic())
		})
	})
})
