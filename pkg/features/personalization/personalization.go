package personalization

import (
	"fmt"

	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/openshift/ocm-container/pkg/features"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

const (
	FeatureFlagName = "no-personalization"
	FlagHelpMessage = "Disable personalizations file mounts and environment"

	destinationDir  = "/root/.config/personalizations.d"
	destinationFile = "/root/.config/personalizations.d/personalizations.sh"
)

type config struct {
	Enabled      bool   `mapstructure:"enabled"`
	Source       string `mapstructure:"source"`
	MountOptions string `mapstructure:"mount_options"`
}

func newConfigWithDefaults() *config {
	config := config{}
	config.Enabled = true
	config.MountOptions = "ro"
	return &config
}

func (cfg *config) validate() error {
	if cfg.MountOptions != "ro" && cfg.MountOptions != "rw" {
		return fmt.Errorf("mount_options must be either 'ro' or 'rw', got: %s", cfg.MountOptions)
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
		log.Debugf("personalization disabled via config")
		return false
	}
	if viper.IsSet(FeatureFlagName) {
		log.Debugf("personalization disabled via flag")
		return false
	}
	if !f.userHasConfig {
		log.Debugf("personalization disabled with no config setup")
		return false
	}
	if f.config.Source == "" {
		log.Debugf("personalization disabled with no source")
		return false
	}

	return true
}

func (f *Feature) ExitOnError() bool {
	return false
}

func (f *Feature) Configure() error {
	cfg := newConfigWithDefaults()

	if !viper.IsSet("features.personalization") {
		f.config = cfg
		return nil
	}

	f.userHasConfig = true
	err := viper.UnmarshalKey("features.personalization", &cfg)
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

	// Check if source is a directory or file
	isDir, err := f.isDirectory(f.config.Source)
	if err != nil {
		return opts, fmt.Errorf("error: problem reading personalization directory or file: %v: %v", f.config.Source, err)
	}

	// If source is a directory, mount the whole directory
	if isDir {
		opts.AddVolumeMount(engine.VolumeMount{
			Source:       f.config.Source,
			Destination:  destinationDir,
			MountOptions: f.config.MountOptions,
		})
	} else {
		// If source is a file, mount it as personalizations.sh
		opts.AddVolumeMount(engine.VolumeMount{
			Source:       f.config.Source,
			Destination:  destinationFile,
			MountOptions: f.config.MountOptions,
		})
	}

	return opts, nil
}

func (f *Feature) HandleError(err error) {
	if f.userHasConfig {
		log.Warnf("Error initializing personalization functionality: %v", err)
	}
	log.Debugf("Error initializing personalization functionality: %v", err)
}

// isDirectory checks if the given path is a directory
// returns true if directory, false if file, error if path doesn't exist
func (f *Feature) isDirectory(path string) (bool, error) {
	fileInfo, err := f.afs.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

func init() {
	f := Feature{
		afs: &afero.Afero{Fs: afero.NewOsFs()},
	}
	features.Register("personalization", &f)
}
