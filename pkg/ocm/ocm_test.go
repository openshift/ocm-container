package ocm

import (
	"github.com/openshift-online/ocm-common/pkg/ocm/config"
	sdk "github.com/openshift-online/ocm-sdk-go"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

var _ = Describe("Pkg/OCM/OCM", func() {
	BeforeEach(func() {
		viper.Reset()
		client = nil
	})

	Context("url()", func() {
		It("Returns production URL for 'prod' alias", func() {
			u, err := url("prod")
			Expect(err).ToNot(HaveOccurred())
			Expect(u).To(Equal(productionURL))
		})

		It("Returns production URL for 'production' alias", func() {
			u, err := url("production")
			Expect(err).ToNot(HaveOccurred())
			Expect(u).To(Equal(productionURL))
		})

		It("Returns production URL for 'prd' alias", func() {
			u, err := url("prd")
			Expect(err).ToNot(HaveOccurred())
			Expect(u).To(Equal(productionURL))
		})

		It("Returns staging URL for 'stage' alias", func() {
			u, err := url("stage")
			Expect(err).ToNot(HaveOccurred())
			Expect(u).To(Equal(stagingURL))
		})

		It("Returns staging URL for 'staging' alias", func() {
			u, err := url("staging")
			Expect(err).ToNot(HaveOccurred())
			Expect(u).To(Equal(stagingURL))
		})

		It("Returns staging URL for 'stg' alias", func() {
			u, err := url("stg")
			Expect(err).ToNot(HaveOccurred())
			Expect(u).To(Equal(stagingURL))
		})

		It("Returns integration URL for 'int' alias", func() {
			u, err := url("int")
			Expect(err).ToNot(HaveOccurred())
			Expect(u).To(Equal(integrationURL))
		})

		It("Returns integration URL for 'integration' alias", func() {
			u, err := url("integration")
			Expect(err).ToNot(HaveOccurred())
			Expect(u).To(Equal(integrationURL))
		})

		It("Returns production gov URL for 'prodgov' alias", func() {
			u, err := url("prodgov")
			Expect(err).ToNot(HaveOccurred())
			Expect(u).To(Equal(productionGovURL))
		})

		It("Returns production gov URL for 'productiongov' alias", func() {
			u, err := url("productiongov")
			Expect(err).ToNot(HaveOccurred())
			Expect(u).To(Equal(productionGovURL))
		})

		It("Returns production gov URL for 'prdgov' alias", func() {
			u, err := url("prdgov")
			Expect(err).ToNot(HaveOccurred())
			Expect(u).To(Equal(productionGovURL))
		})

		It("Returns the URL itself when given a full URL", func() {
			u, err := url(productionURL)
			Expect(err).ToNot(HaveOccurred())
			Expect(u).To(Equal(productionURL))
		})

		It("Returns error for invalid alias", func() {
			_, err := url("invalid")
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(errInvalidOcmUrl))
		})

		It("Returns error for empty string", func() {
			_, err := url("")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("alias()", func() {
		It("Returns 'prod' for production URL", func() {
			Expect(alias(productionURL)).To(Equal("prod"))
		})

		It("Returns 'stage' for staging URL", func() {
			Expect(alias(stagingURL)).To(Equal("stage"))
		})

		It("Returns 'int' for integration URL", func() {
			Expect(alias(integrationURL)).To(Equal("int"))
		})

		It("Returns 'prodgov' for production gov URL", func() {
			Expect(alias(productionGovURL)).To(Equal("prodgov"))
		})

		It("Returns empty string for unknown URL", func() {
			Expect(alias("https://unknown.example.com")).To(BeEmpty())
		})
	})

	Context("Error type", func() {
		It("Returns the error string", func() {
			e := Error("test error")
			Expect(e.Error()).To(Equal("test error"))
		})
	})

	Context("GetClient()", func() {
		It("Returns nil when no client is set", func() {
			Expect(GetClient()).To(BeNil())
		})
	})

	Context("CloseClient()", func() {
		It("Returns nil when no client exists", func() {
			err := CloseClient()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("ensureConfigDefaults()", func() {
		It("Fills in defaults for empty config", func() {
			cfg := &config.Config{}
			err := ensureConfigDefaults(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.ClientID).To(Equal(ocmContainerClientId))
			Expect(cfg.TokenURL).To(Equal(sdk.DefaultTokenURL))
			Expect(cfg.Scopes).To(Equal(defaultOcmScopes))
			Expect(cfg.URL).To(Equal(productionURL))
		})

		It("Preserves existing ClientID", func() {
			cfg := &config.Config{ClientID: "custom-client"}
			err := ensureConfigDefaults(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.ClientID).To(Equal("custom-client"))
		})

		It("Preserves existing TokenURL", func() {
			cfg := &config.Config{TokenURL: "https://custom-token-url"}
			err := ensureConfigDefaults(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.TokenURL).To(Equal("https://custom-token-url"))
		})

		It("Preserves existing Scopes", func() {
			customScopes := []string{"openid", "profile"}
			cfg := &config.Config{Scopes: customScopes}
			err := ensureConfigDefaults(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.Scopes).To(Equal(customScopes))
		})

		It("Preserves existing URL", func() {
			cfg := &config.Config{URL: stagingURL}
			err := ensureConfigDefaults(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.URL).To(Equal(stagingURL))
		})

		It("Uses viper ocm-url when URL is empty", func() {
			viper.Set("ocm-url", "stage")
			cfg := &config.Config{}
			err := ensureConfigDefaults(cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.URL).To(Equal(stagingURL))
		})

		It("Returns error for invalid viper ocm-url", func() {
			viper.Set("ocm-url", "invalid-url")
			cfg := &config.Config{}
			err := ensureConfigDefaults(cfg)
			Expect(err).To(HaveOccurred())
		})
	})
})
