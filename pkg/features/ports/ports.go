package ports

import (
	"fmt"
	"strings"

	"github.com/openshift/ocm-container/pkg/features"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Define any defaults here as constants
const (
	FeatureFlagName = "no-ports"
	FlagHelpMessage = "Disables all additional ports "

	portLookupTemplate = `{{(index (index .NetworkSettings.Ports "%d/tcp") 0).HostPort}}`

	defaultConsolePort = 9999
)

// Any internal config needed for the setup of the feature
type config struct {
	Enabled bool `mapstructure:"enabled"`

	Console portConfig `mapstructure:"console"`
}

type portConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Port    int  `mapstructure:"host-port"`
}

// This is where we want to set all of our config defaults. If
// the user doesn't explicitly NEED to set something, set it
// here for them and allow them to overwrite it.
func newConfigWithDefaults() *config {
	cfg := config{}
	cfg.Enabled = true
	cfg.Console = portConfig{
		Enabled: true,
		Port:    defaultConsolePort,
	}
	return &cfg
}

// Validate is where any custom configuration validation logic
// lives. This is where you need to validate your user's input
func (cfg *config) validate() error {
	// TODO - validate that each individual port is within a valid range
	return nil
}

type Feature struct {
	config *config

	// If the user provided a configuration, set that here
	// in case we want to handle initialization errors
	// differently because of that
	userHasConfig bool
}

// Enabled is where we determine whether or not the feature
// is explicitly enabled if opt-in or disabled if opt-out.
func (f *Feature) Enabled() bool {
	if !f.config.Enabled {
		log.Debugf("all ports disabled via config")
	}
	if viper.IsSet(FeatureFlagName) {
		log.Debugf("all ports disabled via flag")
	}
	return f.config.Enabled && !viper.IsSet(FeatureFlagName)
}

// If this feature is required for the functionality of
// ocm-container OR if a configuration error will be
// catastrophic to our user's experience, set this to true.
// Otherwise, if we lose a convenience function but we should
// still allow the user to use the container, then set false.
// In almost all cases, this should be set to false.
func (f *Feature) ExitOnError() bool {
	return false
}

// We want to self-contain the configuration functionality separate
// from the initialization so that we can read in the enabled config
func (f *Feature) Configure() error {
	cfg := newConfigWithDefaults()

	if !viper.IsSet("features.ports") {
		f.config = cfg
		return nil
	}

	f.userHasConfig = true
	err := viper.UnmarshalKey("features.ports", &cfg)
	if err != nil {
		return err
	}

	// The above gets all of the NAMED ports - if a user wants to
	// define additional ports we'll want to check for those to

	// TODO: check for additional ports in the config

	f.config = cfg
	err = cfg.validate()
	if err != nil {
		return err
	}

	return nil
}

// Initialize is the feature that we use to create the OptionSet
// for a given feature. An OptionSet is how the ocm-container
// program knows what options to pass into the container create
// command in order for the individual feature to work properly
func (f *Feature) Initialize() (features.OptionSet, error) {
	opts := features.NewOptionSet()

	ports := map[string]int{}
	ports["console"] = f.config.Console.Port

	opts.RegisterPortMap(ports)

	opts.RegisterPostStartExecHook(func(o features.ContainerRuntime) error {
		consolePortLookup := fmt.Sprintf(portLookupTemplate, f.config.Console.Port)
		log.Debugf("Inspect for console port")
		consolePort, err := o.Inspect(consolePortLookup)
		if err != nil {
			return err
		}

		portMapCmd := []string{
			"/bin/bash",
			"-c",
			fmt.Sprintf("echo \"%v\" > /tmp/portmap", (consolePort)),
		}

		o.RegisterBlockingPostStartCmd(portMapCmd)
		log.Debugf("Console port blocking command registered: '%s'", strings.Join(portMapCmd, " "))
		return nil
	})

	return opts, nil
}

// If initialize fails, how should we handle the error? This
// allows you to customize what log level to use or how to
// clean up anything you need to.
func (f *Feature) HandleError(err error) {
	// example how we want to handle feature intilization
	// errors differently based on whether or not the user
	// provided a configuration. In this case, if the user
	// provided config themselves we want to explicitly warn
	// them of the error, but if they didn't set a config and
	// the default functionality isn't working, there's no
	// need to inform them because they didn't set it up.
	if f.userHasConfig {
		log.Warnf("Error initializing ports functionality: %v", err)
		return
	}
	log.Debugf("Error initializing ports functionality: %v", err)
}

func init() {
	f := Feature{}
	features.Register("ports", &f)
}
