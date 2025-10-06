package pagerduty

import (
	"fmt"
	"os"
	"slices"

	"github.com/adrg/xdg"
	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/openshift/ocm-container/pkg/features"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	defaultPagerDutyTokenFile = ".config/pagerduty-cli/config.json"
	pagerDutyTokenDest        = "/root/" + defaultPagerDutyTokenFile
)

type Config struct {
	Token string `mapstructure:"token_file"`
	Mount string `mapstructure:"mount"`
}

func newConfigWithDefaults() *Config {
	config := Config{}
	config.Token = defaultPagerDutyTokenFile
	config.Mount = "rw"

	return &config
}

func (cfg *Config) validate() error {
	validMountOptions := []string{
		"ro",
		"rw",
	}
	if !slices.Contains(validMountOptions, cfg.Mount) {
		return fmt.Errorf("invalid mount option. Valid options are %s", validMountOptions)
	}
	return nil
}

func New() (features.OptionSet, error) {
	opts := features.NewOptionSet()

	cfg := newConfigWithDefaults()

	viper.UnmarshalKey("features.pagerduty", &cfg)
	err := cfg.validate()
	if err != nil {
		return opts, err
	}

	pdConfigFile, err := statConfigFileLocations(cfg.Token)
	if err != nil {
		return opts, err
	}

	opts.AddVolumeMount(engine.VolumeMount{
		Source:       pdConfigFile,
		Destination:  pagerDutyTokenDest,
		MountOptions: cfg.Mount,
	})

	return opts, nil
}

// check for config file locations in the following order:
// absolute path -> $HOME/(path) -> xdgConfig/(path)
// return error if not found after all three have been checked
func statConfigFileLocations(filepath string) (string, error) {
	errorPaths := []string{}
	_, err := os.Stat(filepath)
	if err == nil {
		log.Debugf("using %s for PD config", filepath)
		return filepath, nil
	}
	errorPaths = append(errorPaths, filepath)

	path := os.Getenv("HOME") + "/" + filepath
	_, err = os.Stat(path)
	if err == nil {
		log.Debugf("using %s for PD config", path)
		return path, nil
	}
	errorPaths = append(errorPaths, path)

	xdgConfigFile, _ := xdg.ConfigFile(filepath)
	configFilePath, err := xdg.SearchConfigFile(xdgConfigFile)
	if err == nil {
		log.Debugf("using %s for PD config", configFilePath)
		return configFilePath, nil
	}
	errorPaths = append(errorPaths, xdgConfigFile)
	return "", fmt.Errorf("could not find %s in any of: %s", filepath, errorPaths)
}
