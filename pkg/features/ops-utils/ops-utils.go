package opsutils

import (
	"fmt"

	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/openshift/ocm-container/pkg/features"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

const (
	FeatureFlagName = "no-ops-utils"
	FlagHelpMessage = "Disable ops-utils mounts and environment"

	destinationDir      = "/root/ops-utils"
	defaultMountOptions = "ro"
)

type config struct {
	Enabled      bool   `mapstructure:"enabled"`
	SourceDir    string `mapstructure:"source_dir"`
	MountOptions string `mapstructure:"mount_options"`
}

func newConfigWithDefaults() *config {
	config := config{}
	config.Enabled = true
	config.MountOptions = defaultMountOptions
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
		log.Debugf("ops-utils disabled via config")
		return false
	}
	if viper.IsSet(FeatureFlagName) {
		log.Debugf("ops-utils disabled via flag")
		return false
	}
	if !f.userHasConfig {
		log.Debugf("ops-utils disabled with no config setup")
		return false
	}
	if f.config.SourceDir == "" {
		log.Debugf("ops-utils disabled with no source dir")
		return false
	}

	return true
}

func (f *Feature) ExitOnError() bool {
	return true
}

func (f *Feature) Configure() error {
	cfg := newConfigWithDefaults()

	if !viper.IsSet("features.ops_utils") {
		f.config = cfg
		return nil
	}

	f.userHasConfig = true
	err := viper.UnmarshalKey("features.ops_utils", &cfg)
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

	// Validate source directory exists
	_, err := f.afs.Stat(f.config.SourceDir)
	if err != nil {
		return opts, fmt.Errorf("error: problem reading Ops Utils directory: %v: %v", f.config.SourceDir, err)
	}

	// TODO: MountOptions doesn't work for SELinux-enabled systems.
	// Included for feature compatibility with previous version, but should be modified for :Z or other solution
	opts.AddVolumeMount(engine.VolumeMount{
		Source:       f.config.SourceDir,
		Destination:  destinationDir,
		MountOptions: f.config.MountOptions,
	})

	return opts, nil
}

func (f *Feature) HandleError(err error) {
	if f.userHasConfig {
		log.Warnf("Error initializing ops-utils functionality: %v", err)
	}
	log.Debugf("Error initializing ops-utils functionality: %v", err)
}

func init() {
	f := Feature{
		afs: &afero.Afero{Fs: afero.NewOsFs()},
	}
	features.Register("ops-utils", &f)
}
