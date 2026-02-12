package personalization

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

var _ = Describe("Pkg/Features/Personalization/Personalization", func() {
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
			Expect(f.config.MountOptions).To(Equal("ro"))
		})

		It("Overwrites defaults with viper config", func() {
			viper.Set("features.personalization", map[string]any{
				"enabled":       false,
				"source":        "/custom/personalization",
				"mount_options": "rw",
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).To(BeNil())

			Expect(f.userHasConfig).To(BeTrue())
			Expect(f.config.Enabled).To(BeFalse())
			Expect(f.config.Source).To(Equal("/custom/personalization"))
			Expect(f.config.MountOptions).To(Equal("rw"))
		})

		It("Returns an error when viper cannot unmarshal the config", func() {
			viper.Set("features.personalization", map[string]any{
				"enabled": "someString",
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("decoding failed"))
		})
	})

	Context("Tests config.validate()", func() {
		It("Returns nil for valid config with ro mount options", func() {
			cfg := config{
				Enabled:      true,
				Source:       "/some/path",
				MountOptions: "ro",
			}
			err := cfg.validate()
			Expect(err).To(BeNil())
		})

		It("Returns nil for valid config with rw mount options", func() {
			cfg := config{
				Enabled:      true,
				Source:       "/some/path",
				MountOptions: "rw",
			}
			err := cfg.validate()
			Expect(err).To(BeNil())
		})

		It("Returns error for invalid mount options", func() {
			cfg := config{
				Enabled:      true,
				Source:       "/some/path",
				MountOptions: "invalid",
			}
			err := cfg.validate()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("mount_options must be either 'ro' or 'rw'"))
		})

		It("Returns nil for disabled config", func() {
			cfg := config{
				Enabled:      false,
				MountOptions: "ro",
			}
			err := cfg.validate()
			Expect(err).To(BeNil())
		})
	})

	Context("Tests Feature.Enabled()", func() {
		It("Returns true when config is enabled, has user config, and source is set", func() {
			f := Feature{
				config:        &config{Enabled: true, Source: "/some/path"},
				userHasConfig: true,
			}
			Expect(f.Enabled()).To(BeTrue())
		})

		It("Returns false when config is disabled", func() {
			f := Feature{
				config:        &config{Enabled: false, Source: "/some/path"},
				userHasConfig: true,
			}
			Expect(f.Enabled()).To(BeFalse())
		})

		It("Returns false when feature flag is set", func() {
			viper.Set(FeatureFlagName, true)
			f := Feature{
				config:        &config{Enabled: true, Source: "/some/path"},
				userHasConfig: true,
			}
			Expect(f.Enabled()).To(BeFalse())
		})

		It("Returns false when user has no config", func() {
			f := Feature{
				config:        &config{Enabled: true, Source: "/some/path"},
				userHasConfig: false,
			}
			Expect(f.Enabled()).To(BeFalse())
		})

		It("Returns false when source is empty", func() {
			f := Feature{
				config:        &config{Enabled: true, Source: ""},
				userHasConfig: true,
			}
			Expect(f.Enabled()).To(BeFalse())
		})

		It("Returns false when both config is disabled and flag is set", func() {
			viper.Set(FeatureFlagName, true)
			f := Feature{
				config:        &config{Enabled: false, Source: "/some/path"},
				userHasConfig: true,
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

	Context("Tests Feature.Initialize() with directory", func() {
		It("Returns OptionSet with directory mount when source is a directory", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			sourcePath := "/test/personalizations"

			err := afs.MkdirAll(sourcePath, 0755)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:      true,
					Source:       sourcePath,
					MountOptions: "ro",
				},
				userHasConfig: true,
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(1))
			Expect(opts.Mounts[0].Source).To(Equal(sourcePath))
			Expect(opts.Mounts[0].Destination).To(Equal(destinationDir))
			Expect(opts.Mounts[0].MountOptions).To(Equal("ro"))
		})

		It("Uses rw mount options when specified for directory", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			sourcePath := "/test/personalizations"

			err := afs.MkdirAll(sourcePath, 0755)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:      true,
					Source:       sourcePath,
					MountOptions: "rw",
				},
				userHasConfig: true,
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts[0].MountOptions).To(Equal("rw"))
		})
	})

	Context("Tests Feature.Initialize() with file", func() {
		It("Returns OptionSet with file mount when source is a file", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			sourcePath := "/test/.bashrc"

			err := afs.WriteFile(sourcePath, []byte("#!/bin/bash"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:      true,
					Source:       sourcePath,
					MountOptions: "ro",
				},
				userHasConfig: true,
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(1))
			Expect(opts.Mounts[0].Source).To(Equal(sourcePath))
			Expect(opts.Mounts[0].Destination).To(Equal(destinationFile))
			Expect(opts.Mounts[0].MountOptions).To(Equal("ro"))
		})

		It("Uses rw mount options when specified for file", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			sourcePath := "/test/.bashrc"

			err := afs.WriteFile(sourcePath, []byte("#!/bin/bash"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:      true,
					Source:       sourcePath,
					MountOptions: "rw",
				},
				userHasConfig: true,
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts[0].MountOptions).To(Equal("rw"))
		})
	})

	Context("Tests Feature.Initialize() errors", func() {
		It("Returns error when source does not exist", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:      true,
					Source:       "/nonexistent/path",
					MountOptions: "ro",
				},
				userHasConfig: true,
			}

			opts, err := f.Initialize()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("problem reading personalization"))
			Expect(opts.Mounts).To(HaveLen(0))
		})
	})

	Context("Tests isDirectory()", func() {
		It("Returns true when path is a directory", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			dirPath := "/test/dir"

			err := afs.MkdirAll(dirPath, 0755)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
			}

			isDir, err := f.isDirectory(dirPath)
			Expect(err).To(BeNil())
			Expect(isDir).To(BeTrue())
		})

		It("Returns false when path is a file", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			filePath := "/test/file.txt"

			err := afs.WriteFile(filePath, []byte("content"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
			}

			isDir, err := f.isDirectory(filePath)
			Expect(err).To(BeNil())
			Expect(isDir).To(BeFalse())
		})

		It("Returns error when path does not exist", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}

			f := Feature{
				afs: &afs,
			}

			isDir, err := f.isDirectory("/nonexistent")
			Expect(err).ToNot(BeNil())
			Expect(isDir).To(BeFalse())
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
