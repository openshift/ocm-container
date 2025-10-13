package legacyawscredentials

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

var _ = Describe("Pkg/Features/LegacyAwsCredentials/LegacyAwsCredentials", func() {
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
		})

		It("Overwrites defaults with viper config", func() {
			viper.Set("features.legacy_aws_credentials", map[string]any{
				"enabled": false,
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).To(BeNil())

			Expect(f.userHasConfig).To(BeTrue())
			Expect(f.config.Enabled).To(BeFalse())
		})

		It("Returns an error when viper cannot unmarshal the config", func() {
			viper.Set("features.legacy_aws_credentials", map[string]any{
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
			cfg := config{Enabled: true}
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
		It("Returns OptionSet with both AWS files when they exist", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			homeDir := os.Getenv("HOME")
			awsDir := homeDir + "/.aws"

			err := afs.MkdirAll(awsDir, 0755)
			Expect(err).To(BeNil())

			credentialsFile := awsDir + "/credentials"
			configFile := awsDir + "/config"
			err = afs.WriteFile(credentialsFile, []byte("[default]\naws_access_key_id=test"), 0644)
			Expect(err).To(BeNil())
			err = afs.WriteFile(configFile, []byte("[default]\nregion=us-east-1"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled: true,
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(2))
			Expect(opts.Mounts[0].Source).To(Equal(credentialsFile))
			Expect(opts.Mounts[0].Destination).To(Equal("/root/.aws/credentials"))
			Expect(opts.Mounts[0].MountOptions).To(Equal("ro"))
			Expect(opts.Mounts[1].Source).To(Equal(configFile))
			Expect(opts.Mounts[1].Destination).To(Equal("/root/.aws/config"))
			Expect(opts.Mounts[1].MountOptions).To(Equal("ro"))
		})

		It("Returns OptionSet with only credentials file when config doesn't exist", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			homeDir := os.Getenv("HOME")
			awsDir := homeDir + "/.aws"

			err := afs.MkdirAll(awsDir, 0755)
			Expect(err).To(BeNil())

			credentialsFile := awsDir + "/credentials"
			err = afs.WriteFile(credentialsFile, []byte("[default]\naws_access_key_id=test"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled: true,
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(1))
			Expect(opts.Mounts[0].Source).To(Equal(credentialsFile))
			Expect(opts.Mounts[0].Destination).To(Equal("/root/.aws/credentials"))
			Expect(opts.Mounts[0].MountOptions).To(Equal("ro"))
		})

		It("Returns OptionSet with only config file when credentials doesn't exist", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			homeDir := os.Getenv("HOME")
			awsDir := homeDir + "/.aws"

			err := afs.MkdirAll(awsDir, 0755)
			Expect(err).To(BeNil())

			configFile := awsDir + "/config"
			err = afs.WriteFile(configFile, []byte("[default]\nregion=us-east-1"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled: true,
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(1))
			Expect(opts.Mounts[0].Source).To(Equal(configFile))
			Expect(opts.Mounts[0].Destination).To(Equal("/root/.aws/config"))
			Expect(opts.Mounts[0].MountOptions).To(Equal("ro"))
		})

		It("Returns empty OptionSet when neither file exists", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled: true,
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts).To(HaveLen(0))
		})

		It("Returns error when HOME is not set", func() {
			originalHome := os.Getenv("HOME")
			os.Unsetenv("HOME")
			defer os.Setenv("HOME", originalHome)

			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			f := Feature{
				afs: &afs,
				config: &config{
					Enabled: true,
				},
			}

			opts, err := f.Initialize()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("$HOME is not set"))
			Expect(opts.Mounts).To(HaveLen(0))
		})

		It("Mounts files with ro option", func() {
			afs := afero.Afero{Fs: afero.NewMemMapFs()}
			homeDir := os.Getenv("HOME")
			awsDir := homeDir + "/.aws"

			err := afs.MkdirAll(awsDir, 0755)
			Expect(err).To(BeNil())

			credentialsFile := awsDir + "/credentials"
			err = afs.WriteFile(credentialsFile, []byte("[default]\naws_access_key_id=test"), 0644)
			Expect(err).To(BeNil())

			f := Feature{
				afs: &afs,
				config: &config{
					Enabled: true,
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.Mounts[0].MountOptions).To(Equal("ro"))
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
