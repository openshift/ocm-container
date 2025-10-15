package backplane

import (
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
	Enabled bool `mapstructure:"enabled"`
}

func newConfigWithDefaults() *config {
	config := config{}
	config.Enabled = true
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
	b := os.Getenv("BACKPLANE_CONFIG")
	if b == "" {
		home := os.Getenv("HOME")
		b = home + "/" + defaultBackplaneConfig
	}

	_, err := f.afs.Stat(b)
	if err != nil {
		return opts, err
	}

	// Set the BACKPLANE_CONFIG environment variable to the in-container path
	opts.AddEnvKeyVal("BACKPLANE_CONFIG", backplaneConfigDest)

	opts.AddVolumeMount(engine.VolumeMount{
		Source:       b,
		Destination:  backplaneConfigDest,
		MountOptions: backplaneConfigMountOpts,
	})

	return opts, nil
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
