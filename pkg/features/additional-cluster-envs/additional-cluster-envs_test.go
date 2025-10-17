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
			viper.Set("features.additional_cluster_envs", map[string]any{
				"enabled": false,
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).To(BeNil())

			Expect(f.userHasConfig).To(BeTrue())
			Expect(f.config.Enabled).To(BeFalse())
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
		It("Returns true when config is enabled and cluster-id is set", func() {
			viper.Set("cluster-id", "test-cluster-123")
			f := Feature{
				config: &config{Enabled: true},
			}
			Expect(f.Enabled()).To(BeTrue())
		})

		It("Returns false when cluster-id is not set", func() {
			f := Feature{
				config: &config{Enabled: true},
			}
			Expect(f.Enabled()).To(BeFalse())
		})

		It("Returns false when config is disabled", func() {
			viper.Set("cluster-id", "test-cluster-123")
			f := Feature{
				config: &config{Enabled: false},
			}
			Expect(f.Enabled()).To(BeFalse())
		})

		It("Returns false when feature flag is set", func() {
			viper.Set("cluster-id", "test-cluster-123")
			viper.Set(FeatureFlagName, true)
			f := Feature{
				config: &config{Enabled: true},
			}
			Expect(f.Enabled()).To(BeFalse())
		})

		It("Returns false when both config is disabled and flag is set", func() {
			viper.Set("cluster-id", "test-cluster-123")
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
		// Note: Initialize() requires OCM client which we can't easily mock in unit tests
		// The Initialize() method is tested through integration tests
		It("Would require OCM client mocking for proper unit testing", func() {
			// This is a placeholder to acknowledge that Initialize() needs integration testing
			// since it calls ocm.NewClient() and ocm.GetCluster()
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
