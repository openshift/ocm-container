package imagecache

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

var _ = Describe("Pkg/Features/ImageCache/ImageCache", func() {
	BeforeEach(func() {
		viper.Reset()
	})

	Context("Tests the config", func() {
		It("Builds the defaults correctly", func() {
			cfg := newConfigWithDefaults()
			Expect(cfg).ToNot(BeNil())
			Expect(cfg.Enabled).To(BeFalse())
			Expect(cfg.StorageDir).To(Equal(defaultStorageDir))
		})
	})

	Context("Tests config.validate()", func() {
		It("Returns nil for valid config", func() {
			cfg := config{
				Enabled:    true,
				StorageDir: "/some/path",
			}
			err := cfg.validate()
			Expect(err).To(BeNil())
		})

		It("Returns nil for disabled config", func() {
			cfg := config{
				Enabled:    false,
				StorageDir: "",
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
			Expect(f.config.Enabled).To(BeFalse())
			Expect(f.config.StorageDir).To(Equal(defaultStorageDir))
		})

		It("Overwrites defaults with viper config", func() {
			viper.Set("features.image_cache", map[string]any{
				"enabled":     true,
				"storage_dir": "/custom/storage",
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).To(BeNil())

			Expect(f.userHasConfig).To(BeTrue())
			Expect(f.config.Enabled).To(BeTrue())
			Expect(f.config.StorageDir).To(Equal("/custom/storage"))
		})

		It("Returns an error when viper cannot unmarshal the config", func() {
			viper.Set("features.image_cache", map[string]any{
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
				config: &config{Enabled: true, StorageDir: "/some/path"},
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
				config: &config{Enabled: true, StorageDir: "/some/path"},
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
		It("Returns false when criticalError is false", func() {
			f := Feature{criticalError: false}
			Expect(f.ExitOnError()).To(BeFalse())
		})

		It("Returns true when criticalError is true", func() {
			f := Feature{criticalError: true}
			Expect(f.ExitOnError()).To(BeTrue())
		})
	})

	Context("Tests statStorageDir()", func() {
		It("Returns absolute path when it exists and is a directory", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			storagePath := "/absolute/storage"

			err := afs.MkdirAll(storagePath, 0755)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					StorageDir: storagePath,
				},
			}

			result, err := f.statStorageDir()
			Expect(err).To(BeNil())
			Expect(result).To(Equal(storagePath))
		})

		It("Returns error and sets criticalError when absolute path exists but is not a directory", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			filePath := "/absolute/file"

			err := afs.WriteFile(filePath, []byte("content"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					StorageDir: filePath,
				},
			}

			result, err := f.statStorageDir()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("not a directory"))
			Expect(result).To(Equal(""))
			Expect(f.criticalError).To(BeTrue())
		})

		It("Returns HOME-relative path when it exists and is a directory", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			homeDir := os.Getenv("HOME")
			relativeDir := ".config/test-storage"
			fullPath := homeDir + "/" + relativeDir

			err := afs.MkdirAll(fullPath, 0755)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					StorageDir: relativeDir,
				},
			}

			result, err := f.statStorageDir()
			Expect(err).To(BeNil())
			Expect(result).To(Equal(fullPath))
		})

		It("Returns error and sets criticalError when HOME-relative path exists but is not a directory", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			homeDir := os.Getenv("HOME")
			relativeFile := ".config/test-file"
			fullPath := homeDir + "/" + relativeFile

			err := afs.WriteFile(fullPath, []byte("content"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					StorageDir: relativeFile,
				},
			}

			result, err := f.statStorageDir()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("not a directory"))
			Expect(result).To(Equal(""))
			Expect(f.criticalError).To(BeTrue())
		})

		It("Returns error when directory doesn't exist in any location", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			relativeDir := ".config/nonexistent"

			f := Feature{
				afs: &afs,
				config: &config{
					StorageDir: relativeDir,
				},
			}

			result, err := f.statStorageDir()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("could not find"))
			Expect(result).To(Equal(""))
		})

		It("Returns error when storage directory is empty", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}

			f := Feature{
				afs: &afs,
				config: &config{
					StorageDir: "",
				},
			}

			result, err := f.statStorageDir()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("no storage directory provided"))
			Expect(result).To(Equal(""))
		})
	})

	Context("Tests Feature.Initialize()", func() {
		It("Creates volume mount with correct destination and options when directory exists", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			homeDir := os.Getenv("HOME")
			relativeDir := ".config/image-cache"
			expectedPath := homeDir + "/" + relativeDir

			// Create the directory first
			err := afs.MkdirAll(expectedPath, 0755)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:    true,
					StorageDir: relativeDir,
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(1))
			Expect(opts.Mounts[0].Source).To(Equal(expectedPath))
			Expect(opts.Mounts[0].Destination).To(Equal(destDir))
			Expect(opts.Mounts[0].MountOptions).To(Equal("rw"))
		})

		It("Returns error when storage directory doesn't exist", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			homeDir := os.Getenv("HOME")
			relativeDir := ".config/nonexistent-cache"
			expectedPath := homeDir + "/" + relativeDir

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:    true,
					StorageDir: relativeDir,
				},
			}

			// Verify directory doesn't exist
			_, err := afs.Stat(expectedPath)
			Expect(err).ToNot(BeNil())

			opts, err := f.Initialize()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("error locating storage directory"))
			Expect(opts.Mounts).To(HaveLen(0))
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
