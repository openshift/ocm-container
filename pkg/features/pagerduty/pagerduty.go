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

type config struct {
	token string `mapstructure:"token_file"`
	mount string `mapstructure:"mount"`
}

func newConfigWithDefaults() config {
	config := config{}
	config.token = defaultPagerDutyTokenFile
	config.mount = "rw"

	return config
}

func (cfg *config) validate() error {
	validMountOptions := []string{
		"ro",
		"rw",
	}
	if !slices.Contains(validMountOptions, cfg.mount) {
		return fmt.Errorf("invalid mount option. Valid options are %s", validMountOptions)
	}
	return nil
}

func New() (features.OptionSet, error) {
	opts := features.NewOptionSet()

	config := newConfigWithDefaults()
	viper.UnmarshalKey("feature.pagerduty", &config)
	log.Debugf("%+v", config)
	err := config.validate()
	if err != nil {
		return opts, err
	}

	pdConfigFile, err := statConfigFileLocations(config.token)
	if err != nil {
		return opts, err
	}

	opts.AddVolumeMount(engine.VolumeMount{
		Source:       pdConfigFile,
		Destination:  pagerDutyTokenDest,
		MountOptions: config.mount,
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

	configFilePath, err := xdg.SearchConfigFile(filepath)
	if err == nil {
		log.Debugf("using %s for PD config", configFilePath)
		return configFilePath, nil
	}
	errorPaths = append(errorPaths, configFilePath)
	return "", fmt.Errorf("could not find %s in any of: %s", filepath, errorPaths)
}
