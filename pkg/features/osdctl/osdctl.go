package osdctl

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
	FeatureFlagName = "no-osdctl"
	FlagHelpMessage = "Disable OSDCTL mounts and environment"

	osdctlConfigFile = ".config/osdctl"
	vaultTokenFile   = ".vault-token"
)

type config struct {
	Enabled            bool   `mapstructure:"enabled"`
	ConfigFile         string `mapstructure:"config_file"`
	TokenFile          string `mapstructure:"token_file"`
	ConfigMountOptions string `mapstructure:"config_mount_options"`
	TokenMountOptions  string `mapstructure:"token_mount_options"`
}

func newConfigWithDefaults() *config {
	config := config{}
	config.Enabled = true
	config.ConfigFile = osdctlConfigFile
	config.TokenFile = vaultTokenFile
	config.ConfigMountOptions = "ro"
	config.TokenMountOptions = "rw"
	return &config
}

func (cfg *config) validate() error {
	if cfg.ConfigMountOptions != "ro" && cfg.ConfigMountOptions != "rw" {
		return fmt.Errorf("config_mount_options must be either 'ro' or 'rw', got: %s", cfg.ConfigMountOptions)
	}
	if cfg.TokenMountOptions != "ro" && cfg.TokenMountOptions != "rw" {
		return fmt.Errorf("token_mount_options must be either 'ro' or 'rw', got: %s", cfg.TokenMountOptions)
	}
	return nil
}

type Feature struct {
	config *config

	userHasConfig bool
	afs           *afero.Afero
	criticalError bool
}

func (f *Feature) Enabled() bool {
	if !f.config.Enabled {
		log.Debugf("osdctl disabled via config")
		return false
	}
	if viper.IsSet(FeatureFlagName) {
		log.Debugf("osdctl disabled via flag")
		return false
	}
	if !f.userHasConfig {
		log.Debugf("osdctl disabled with no config setup")
		return false
	}
	if f.config.ConfigFile == "" {
		log.Debugf("osdctl disabled with no config file")
		return false
	}

	return true
}

func (f *Feature) ExitOnError() bool {
	return f.criticalError
}

func (f *Feature) Configure() error {
	cfg := newConfigWithDefaults()

	if !viper.IsSet("features.osdctl") {
		f.config = cfg
		return nil
	}

	f.userHasConfig = true
	err := viper.UnmarshalKey("features.osdctl", &cfg)
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

	configPath, err := f.statFileLocations(f.config.ConfigFile)
	if err != nil {
		log.Infof("Could not find osdctl config file: %v", err)
		// we only want to create a critical error here if the user
		// has set up the config, otherwise we just fail silently
		if f.userHasConfig {
			f.criticalError = true
		}
		return opts, err
	}

	opts.AddVolumeMount(engine.VolumeMount{
		Source:       configPath,
		Destination:  "/root/" + osdctlConfigFile,
		MountOptions: f.config.ConfigMountOptions,
	})

	// Optionally mount the vault token file if it exists
	tokenPath, err := f.statFileLocations(f.config.TokenFile)
	if err == nil {
		opts.AddVolumeMount(engine.VolumeMount{
			Source:       tokenPath,
			Destination:  "/root/" + vaultTokenFile,
			MountOptions: f.config.TokenMountOptions,
		})
	} else {
		log.Debugf("vault token file not found, skipping mount: %v", err)
	}

	return opts, nil
}

func (f *Feature) HandleError(err error) {
	if f.userHasConfig {
		log.Warnf("Error initializing osdctl functionality: %v", err)
	}
	log.Debugf("Error initializing osdctl functionality: %v", err)
}

// check for file locations in the following order:
// absolute path -> $HOME/(path)
// return error if not found after all have been checked
func (f *Feature) statFileLocations(filepath string) (string, error) {
	if filepath == "" {
		return "", fmt.Errorf("no filepath provided")
	}
	errorPaths := []string{}
	_, err := f.afs.Stat(filepath)
	if err == nil {
		log.Debugf("using %s for file", filepath)
		return filepath, nil
	}
	errorPaths = append(errorPaths, filepath)

	path := os.Getenv("HOME") + "/" + filepath
	_, err = f.afs.Stat(path)
	if err == nil {
		log.Debugf("using %s for file", path)
		return path, nil
	}
	errorPaths = append(errorPaths, path)

	return "", fmt.Errorf("could not find %s in any of: %s", filepath, errorPaths)
}

func init() {
	f := Feature{
		afs: &afero.Afero{Fs: afero.NewOsFs()},
	}
	features.Register("osdctl", &f)
}
