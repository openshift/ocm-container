package ports

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

var _ = Describe("Pkg/Features/Ports/Ports", func() {
	BeforeEach(func() {
		viper.Reset()
	})

	Context("Tests the config", func() {
		It("Builds the defaults correctly", func() {
			cfg := newConfigWithDefaults()
			Expect(cfg).ToNot(BeNil())
			Expect(cfg.Enabled).To(BeTrue())
			Expect(cfg.Console.Enabled).To(BeTrue())
			Expect(cfg.Console.Port).To(Equal(defaultConsolePort))
			Expect(cfg.Vault.Enabled).To(BeTrue())
			Expect(cfg.Vault.Port).To(Equal(defaultVaultPort))
		})
	})

	Context("Tests config.validate()", func() {
		It("Returns nil for valid config", func() {
			cfg := config{
				Enabled: true,
				Console: portConfig{
					Enabled: true,
					Port:    9999,
				},
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

		It("Returns nil for disabled console port", func() {
			cfg := config{
				Enabled: true,
				Console: portConfig{
					Enabled: false,
					Port:    0,
				},
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
			Expect(f.config.Console.Enabled).To(BeTrue())
			Expect(f.config.Console.Port).To(Equal(defaultConsolePort))
			Expect(f.config.Vault.Enabled).To(BeTrue())
			Expect(f.config.Vault.Port).To(Equal(defaultVaultPort))
		})

		It("Overwrites defaults with viper config", func() {
			viper.Set("ports", map[string]any{
				"enabled": false,
				"console": map[string]any{
					"enabled": false,
					"port":    8888,
				},
				"vault": map[string]any{
					"enabled": false,
					"port":    9250,
				},
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).To(BeNil())

			Expect(f.userHasConfig).To(BeTrue())
			Expect(f.config.Enabled).To(BeFalse())
			Expect(f.config.Console.Enabled).To(BeFalse())
			Expect(f.config.Console.Port).To(Equal(8888))
			Expect(f.config.Vault.Enabled).To(BeFalse())
			Expect(f.config.Vault.Port).To(Equal(9250))
		})

		It("Overwrites only console port in viper config", func() {
			viper.Set("ports", map[string]any{
				"console": map[string]any{
					"port": 7777,
				},
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).To(BeNil())

			Expect(f.userHasConfig).To(BeTrue())
			Expect(f.config.Enabled).To(BeTrue())
			Expect(f.config.Console.Enabled).To(BeTrue())
			Expect(f.config.Console.Port).To(Equal(7777))
		})

		It("Overwrites only console enabled in viper config", func() {
			viper.Set("ports", map[string]any{
				"console": map[string]any{
					"enabled": false,
				},
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).To(BeNil())

			Expect(f.userHasConfig).To(BeTrue())
			Expect(f.config.Enabled).To(BeTrue())
			Expect(f.config.Console.Enabled).To(BeFalse())
			Expect(f.config.Console.Port).To(Equal(defaultConsolePort))
		})

		It("Returns an error when viper cannot unmarshal the config", func() {
			viper.Set("ports", map[string]any{
				"enabled": "someString",
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("decoding failed"))
		})

		It("Returns an error when console config is invalid", func() {
			viper.Set("ports", map[string]any{
				"console": "invalidString",
			})
			f := Feature{}
			err := f.Configure()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("decoding failed"))
		})

		It("Returns an error when console port is invalid type", func() {
			viper.Set("ports", map[string]any{
				"console": map[string]any{
					"port": "notAnInt",
				},
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
		It("Registers port map with default console and vault ports", func() {
			f := Feature{
				config: &config{
					Enabled: true,
					Console: portConfig{
						Enabled: true,
						Port:    defaultConsolePort,
					},
					Vault: portConfig{
						Enabled: true,
						Port:    defaultVaultPort,
					},
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.PortMap).To(HaveLen(2))
			Expect(opts.PortMap).To(HaveKey("console"))
			Expect(opts.PortMap["console"]).To(Equal(defaultConsolePort))
			Expect(opts.PortMap).To(HaveKey("vault"))
			Expect(opts.PortMap["vault"]).To(Equal(defaultVaultPort))
		})

		It("Registers port map with custom console port", func() {
			customPort := 8888
			f := Feature{
				config: &config{
					Enabled: true,
					Console: portConfig{
						Enabled: true,
						Port:    customPort,
					},
					Vault: portConfig{
						Enabled: true,
						Port:    defaultVaultPort,
					},
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.PortMap).To(HaveLen(2))
			Expect(opts.PortMap).To(HaveKey("console"))
			Expect(opts.PortMap["console"]).To(Equal(customPort))
		})

		It("Registers port map with custom vault port", func() {
			customPort := 9250
			f := Feature{
				config: &config{
					Enabled: true,
					Console: portConfig{
						Enabled: true,
						Port:    defaultConsolePort,
					},
					Vault: portConfig{
						Enabled: true,
						Port:    customPort,
					},
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.PortMap).To(HaveLen(2))
			Expect(opts.PortMap).To(HaveKey("vault"))
			Expect(opts.PortMap["vault"]).To(Equal(customPort))
		})

		It("Registers PostStartExecHooks for console and vault", func() {
			f := Feature{
				config: &config{
					Enabled: true,
					Console: portConfig{
						Enabled: true,
						Port:    defaultConsolePort,
					},
					Vault: portConfig{
						Enabled: true,
						Port:    defaultVaultPort,
					},
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.PostStartExecHooks).To(HaveLen(2))
		})

		It("Registers only console port when vault is disabled", func() {
			f := Feature{
				config: &config{
					Enabled: true,
					Console: portConfig{
						Enabled: false,
						Port:    defaultConsolePort,
					},
					Vault: portConfig{
						Enabled: false,
						Port:    defaultVaultPort,
					},
				},
			}

			opts, err := f.Initialize()
			Expect(err).To(BeNil())
			Expect(opts.PortMap).To(HaveLen(1))
			Expect(opts.PortMap["console"]).To(Equal(defaultConsolePort))
			Expect(opts.PortMap).ToNot(HaveKey("vault"))
			Expect(opts.PostStartExecHooks).To(HaveLen(1))
		})
	})

	Context("Tests Feature.HandleError()", func() {
		It("Does not panic when userHasConfig is true", func() {
			f := Feature{userHasConfig: true}
			Expect(func() {
				f.HandleError(fmt.Errorf("test"))
			}).NotTo(Panic())
		})

		It("Does not panic when userHasConfig is false", func() {
			f := Feature{userHasConfig: false}
			Expect(func() {
				f.HandleError(fmt.Errorf("test"))
			}).NotTo(Panic())
		})
	})
})
