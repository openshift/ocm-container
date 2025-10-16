package backplane

import (
	"fmt"
	"os"

	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/openshift/ocm-container/pkg/features"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

const (
	FeatureFlagName = "no-backplane"
	FlagHelpMessage = "Disable backplane configuration mounting"

	backplaneConfigDest      = "/root/.config/backplane/config.json"
	backplaneConfigMountOpts = "rw"
	defaultBackplaneConfig   = ".config/backplane/config.json"
)

type config struct {
	Enabled    bool   `mapstructure:"enabled"`
	ConfigFile string `mapstructure:"config_file"`
}

func newConfigWithDefaults() *config {
	config := config{}
	config.Enabled = true
	config.ConfigFile = defaultBackplaneConfig
	return &config
}

func (cfg *config) validate() error {
	return nil
}

type Feature struct {
	config *config

	userHasConfig bool
	afs           *afero.Afero
}

func (f *Feature) Enabled() bool {
	if !f.config.Enabled {
		log.Debugf("backplane disabled via config")
		return false
	}
	if viper.IsSet(FeatureFlagName) {
		log.Debugf("backplane disabled via flag")
		return false
	}
	return true
}

func (f *Feature) ExitOnError() bool {
	return false
}

func (f *Feature) Configure() error {
	cfg := newConfigWithDefaults()

	if !viper.IsSet("features.backplane") {
		f.config = cfg
		return nil
	}

	f.userHasConfig = true
	err := viper.UnmarshalKey("features.backplane", &cfg)
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

	// Determine backplane config location
	// Priority: BACKPLANE_CONFIG env var > config_file setting > default
	var configPath string
	backplaneEnv := os.Getenv("BACKPLANE_CONFIG")
	if backplaneEnv != "" {
		// If BACKPLANE_CONFIG is set, use it directly
		configPath = backplaneEnv
	} else {
		// Otherwise use config_file with path resolution
		var err error
		configPath, err = f.statFileLocation(f.config.ConfigFile)
		if err != nil {
			return opts, err
		}
	}

	// Verify the config file exists
	_, err := f.afs.Stat(configPath)
	if err != nil {
		return opts, err
	}

	// Set the BACKPLANE_CONFIG environment variable to the in-container path
	opts.AddEnvKeyVal("BACKPLANE_CONFIG", backplaneConfigDest)

	opts.AddVolumeMount(engine.VolumeMount{
		Source:       configPath,
		Destination:  backplaneConfigDest,
		MountOptions: backplaneConfigMountOpts,
	})

	return opts, nil
}

// statFileLocation checks for file locations in the following order:
// absolute path -> $HOME/(path)
// Returns an error if not found after all have been checked
func (f *Feature) statFileLocation(filepath string) (string, error) {
	if filepath == "" {
		return "", fmt.Errorf("no filepath provided")
	}
	errorPaths := []string{}

	// Check absolute path first
	_, err := f.afs.Stat(filepath)
	if err == nil {
		log.Debugf("using %s for backplane config", filepath)
		return filepath, nil
	}
	errorPaths = append(errorPaths, filepath)

	// Try $HOME-relative path
	path := os.Getenv("HOME") + "/" + filepath
	_, err = f.afs.Stat(path)
	if err == nil {
		log.Debugf("using %s for backplane config", path)
		return path, nil
	}
	errorPaths = append(errorPaths, path)

	return "", fmt.Errorf("could not find %s in any of: %v", filepath, errorPaths)
}

func (f *Feature) HandleError(err error) {
	if f.userHasConfig {
		log.Warnf("Error initializing backplane functionality: %v", err)
	}
	log.Debugf("Error initializing backplane functionality: %v", err)
}

func init() {
	f := Feature{
		afs: &afero.Afero{Fs: afero.NewOsFs()},
	}
	features.Register("backplane", &f)
}
