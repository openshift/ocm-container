package gcloud

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
	FeatureFlagName = "no-gcloud"
	FlagHelpMessage = "Disable GCloud configuration mounting"

	gcloudConfigDir = ".config/gcloud"
)

type config struct {
	Enabled   bool   `mapstructure:"enabled"`
	ConfigDir string `mapstructure:"config_dir"`
	MountOpts string `mapstructure:"config_mount"`
}

func newConfigWithDefaults() *config {
	config := config{}
	config.Enabled = true
	config.ConfigDir = gcloudConfigDir
	config.MountOpts = "ro"
	return &config
}

func (cfg *config) validate() error {
	validMountOptions := []string{
		"ro",
		"rw",
		"z",
		"Z",
		"ro,z",
		"ro,Z",
		"rw,z",
		"rw,Z",
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
		log.Debugf("GCloud disabled via config")
	}
	if viper.IsSet(FeatureFlagName) {
		log.Debugf("GCloud disabled via flag")
	}
	return f.config.Enabled && !viper.IsSet(FeatureFlagName)
}

func (f *Feature) ExitOnError() bool {
	return false
}

func (f *Feature) Configure() error {
	cfg := newConfigWithDefaults()

	if !viper.IsSet("features.gcloud") {
		f.config = cfg
		return nil
	}

	f.userHasConfig = true
	err := viper.UnmarshalKey("features.gcloud", &cfg)
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

	configPath, err := f.statConfigFileLocations()
	if err != nil {
		return opts, err
	}

	opts.AddVolumeMount(engine.VolumeMount{
		Source:       configPath,
		Destination:  "/root/" + gcloudConfigDir,
		MountOptions: f.config.MountOpts,
	})

	return opts, nil
}

func (f *Feature) HandleError(err error) {
	if f.userHasConfig {
		log.Warnf("Error initializing GCloud functionality: %v", err)
	}
	log.Debugf("Error initializing GCloud functionality: %v", err)
}

// check for config file locations in the following order:
// absolute path -> $HOME/(path)
// return error if not found after all have been checked
func (f *Feature) statConfigFileLocations() (string, error) {
	filepath := f.config.ConfigDir
	if filepath == "" {
		return "", fmt.Errorf("no filepath provided")
	}
	errorPaths := []string{}
	_, err := f.afs.Stat(filepath)
	if err == nil {
		log.Debugf("using %s for gcloud config dir", filepath)
		return filepath, nil
	}
	errorPaths = append(errorPaths, filepath)

	path := os.Getenv("HOME") + "/" + filepath
	_, err = f.afs.Stat(path)
	if err == nil {
		log.Debugf("using %s for gcloud config dir", path)
		return path, nil
	}
	errorPaths = append(errorPaths, path)

	return "", fmt.Errorf("could not find %s in any of: %s", filepath, errorPaths)
}

func init() {
	f := Feature{
		afs: &afero.Afero{Fs: afero.NewOsFs()},
	}
	features.Register("gcloud", &f)
}
