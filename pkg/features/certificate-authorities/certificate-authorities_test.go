package certificateauthorities

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

var _ = Describe("Pkg/Features/CertificateAuthorities/CertificateAuthorities", func() {
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
			Expect(f.config.SourcePath).To(Equal(defaultCaTrustSourceAnchorPath))
		})

		It("Overwrites defaults with viper config", func() {
			viper.Set("features.certificate_authorities", map[string]any{
				"enabled":        false,
				"source_anchors": "/custom/ca/path",
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).To(BeNil())

			Expect(f.userHasConfig).To(BeTrue())
			Expect(f.config.Enabled).To(BeFalse())
			Expect(f.config.SourcePath).To(Equal("/custom/ca/path"))
		})

		It("Returns an error when viper cannot unmarshal the config", func() {
			viper.Set("features.certificate_authorities", map[string]any{
				"enabled": "someString",
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("decoding failed"))
		})
	})

	Context("Tests config.validate()", func() {
		It("Returns nil for valid config", func() {
			cfg := config{Enabled: true, SourcePath: "/some/path"}
			err := cfg.validate()
			Expect(err).To(BeNil())
		})

		It("Returns nil for disabled config", func() {
			cfg := config{Enabled: false}
			err := cfg.validate()
			Expect(err).To(BeNil())
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
		It("Returns OptionSet with volume mount when CA path exists", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			caPath := "/etc/pki/ca-trust/source/anchors"

			err := afs.MkdirAll(caPath, 0755)
			Expect(err).To(BeNil())
			err = afs.WriteFile(caPath+"/my-ca.crt", []byte("cert content"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:    true,
					SourcePath: caPath,
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(1))
			Expect(opts.Mounts[0].Source).To(Equal(caPath))
			Expect(opts.Mounts[0].Destination).To(Equal(defaultcaTrustDestinationPath))
			Expect(opts.Mounts[0].MountOptions).To(Equal("ro"))
		})

		It("Returns OptionSet with custom source path when specified", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			customPath := "/custom/ca/path"

			err := afs.MkdirAll(customPath, 0755)
			Expect(err).To(BeNil())
			err = afs.WriteFile(customPath+"/custom-ca.crt", []byte("cert content"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:    true,
					SourcePath: customPath,
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(1))
			Expect(opts.Mounts[0].Source).To(Equal(customPath))
			Expect(opts.Mounts[0].Destination).To(Equal(defaultcaTrustDestinationPath))
		})

		It("Returns error when CA path does not exist", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:    true,
					SourcePath: "/nonexistent/path",
				},
			}

			opts, err := f.Initialize()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("problem reading CA Anchors"))
			Expect(opts.Mounts).To(HaveLen(0))
		})

		It("Mounts with ro option", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			caPath := "/etc/pki/ca-trust/source/anchors"

			err := afs.MkdirAll(caPath, 0755)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:    true,
					SourcePath: caPath,
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts[0].MountOptions).To(Equal("ro"))
		})

		It("Uses default destination path", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			caPath := "/custom/source"

			err := afs.MkdirAll(caPath, 0755)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled:    true,
					SourcePath: caPath,
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts[0].Destination).To(Equal(defaultcaTrustDestinationPath))
		})
	})

	Context("Tests Feature.HandleError()", func() {
		It("Does not panic when userHasConfig is true", func() {
			f := Feature{userHasConfig: true}
			Expect(func() {
				f.HandleError(afero.ErrFileNotFound)
			}).NotTo(Panic())
		})

		It("Does not panic when userHasConfig is false", func() {
			f := Feature{userHasConfig: false}
			Expect(func() {
				f.HandleError(afero.ErrFileNotFound)
			}).NotTo(Panic())
		})
	})
})
