package additionalclusterenvs

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

var _ = Describe("Pkg/Features/AdditionalClusterEnvs/AdditionalClusterEnvs", func() {
	BeforeEach(func() {
		viper.Reset()
	})

	Context("Tests the config", func() {
		It("Builds the defaults correctly", func() {
			cfg := newConfigWithDefaults()
			Expect(cfg).ToNot(BeNil())
			Expect(cfg.Enabled).To(BeTrue())
			Expect(cfg.EnvVars).ToNot(BeNil())
			Expect(cfg.EnvVars).To(BeEmpty())
		})
	})

	Context("Tests config.validate()", func() {
		It("Returns nil for valid config", func() {
			cfg := config{
				Enabled: true,
				EnvVars: map[string]string{
					"MY_VAR": "my_value",
				},
			}
			err := cfg.validate()
			Expect(err).To(BeNil())
		})

		It("Returns nil for disabled config", func() {
			cfg := config{
				Enabled: false,
				EnvVars: nil,
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
			Expect(f.config.EnvVars).ToNot(BeNil())
		})

		It("Overwrites defaults with viper config", func() {
			viper.Set("features.additional_cluster_envs", map[string]any{
				"enabled": false,
				"env_vars": map[string]string{
					"CUSTOM_VAR": "custom_value",
				},
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).To(BeNil())

			Expect(f.userHasConfig).To(BeTrue())
			Expect(f.config.Enabled).To(BeFalse())
			Expect(f.config.EnvVars).To(HaveKeyWithValue("CUSTOM_VAR", "custom_value"))
		})

		It("Returns an error when viper cannot unmarshal the config", func() {
			viper.Set("features.additional_cluster_envs", map[string]any{
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
		It("Returns empty OptionSet when no env vars are configured", func() {
			f := Feature{
				config: &config{
					Enabled: true,
					EnvVars: map[string]string{},
				},
			}
			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts).ToNot(BeNil())
		})

		It("Adds environment variables to OptionSet", func() {
			f := Feature{
				config: &config{
					Enabled: true,
					EnvVars: map[string]string{
						"VAR1": "value1",
						"VAR2": "value2",
					},
				},
			}
			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts).ToNot(BeNil())
		})
	})

	Context("Tests Feature.HandleError()", func() {
		It("Does not panic when userHasConfig is true", func() {
			f := Feature{userHasConfig: true}
			Expect(func() {
				f.HandleError(nil)
			}).NotTo(Panic())
		})

		It("Does not panic when userHasConfig is false", func() {
			f := Feature{userHasConfig: false}
			Expect(func() {
				f.HandleError(nil)
			}).NotTo(Panic())
		})
	})
})
