package pagerduty

import (
	"fmt"
	"os"
	"slices"

	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/openshift/ocm-container/pkg/features"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

const (
	FeatureFlagName = "no-pagerduty"
	FlagHelpMessage = "Disable PagerDuty CLI mounts and environment"

	defaultPagerDutyTokenFile = ".config/pagerduty-cli/config.json"
	pagerDutyTokenDest        = "/root/" + defaultPagerDutyTokenFile
)

type config struct {
	Enabled   bool   `mapstructure:"enabled"`
	FilePath  string `mapstructure:"config_file"`
	MountOpts string `mapstructure:"config_mount"`
}

func newConfigWithDefaults() *config {
	config := config{}
	config.Enabled = true
	config.FilePath = defaultPagerDutyTokenFile
	config.MountOpts = "rw"

	return &config
}

func (cfg *config) validate() error {
	validMountOptions := []string{
		"ro",
		"rw",
	}
	if !slices.Contains(validMountOptions, cfg.MountOpts) {
		return fmt.Errorf("invalid mount option. Valid options are %s", validMountOptions)
	}
	return nil
}

type Feature struct {
	config *config

	userHasConfig bool
	afs           *afero.Afero
}

func (f *Feature) Enabled() bool {
	if !f.config.Enabled {
		log.Debugf("Pagerduty disabled via config")
	}
	if viper.IsSet(FeatureFlagName) {
		log.Debugf("Pagerduty disabled via flag")
	}
	return f.config.Enabled && !viper.IsSet(FeatureFlagName)
}

func (f *Feature) ExitOnError() bool {
	return false
}

func (f *Feature) Configure() error {
	cfg := newConfigWithDefaults()

	if !viper.IsSet("features.pagerduty") {
		// if they haven't set a config, exit here
		f.config = cfg
		return nil
	}

	f.userHasConfig = true
	err := viper.UnmarshalKey("features.pagerduty", &cfg)
	if err != nil {
		return err
	}

	f.config = cfg
	err = cfg.validate()
	if err != nil {
		return err
	}

	return nil
}

func (f *Feature) Initialize() (features.OptionSet, error) {
	opts := features.NewOptionSet()

	pdConfigFile, err := f.statConfigFileLocations()
	if err != nil {
		return opts, err
	}

	opts.AddVolumeMount(engine.VolumeMount{
		Source:       pdConfigFile,
		Destination:  pagerDutyTokenDest,
		MountOptions: f.config.MountOpts,
	})

	return opts, nil
}

func (f *Feature) HandleError(err error) {
	if f.userHasConfig {
		log.Warnf("Error initializing PagerDuty functionality: %v", err)
	}
	log.Debugf("Error initializing PagerDuty functionality: %v", err)
}

// check for config file locations in the following order:
// absolute path -> $HOME/(path) -> xdgConfig/(path)
// return error if not found after all three have been checked
func (f *Feature) statConfigFileLocations() (string, error) {
	filepath := f.config.FilePath
	errorPaths := []string{}
	_, err := f.afs.Stat(filepath)
	if err == nil {
		log.Debugf("using %s for PD config", filepath)
		return filepath, nil
	}
	errorPaths = append(errorPaths, filepath)

	path := os.Getenv("HOME") + "/" + filepath
	_, err = f.afs.Stat(path)
	if err == nil {
		log.Debugf("using %s for PD config", path)
		return path, nil
	}
	errorPaths = append(errorPaths, path)

	return "", fmt.Errorf("could not find %s in any of: %s", filepath, errorPaths)
}

func init() {
	f := Feature{
		afs: &afero.Afero{Fs: afero.NewOsFs()},
	}
	features.Register("pagerduty", &f)
}
