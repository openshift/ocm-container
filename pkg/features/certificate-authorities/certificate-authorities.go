package certificateauthorities

import (
	"fmt"

	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/openshift/ocm-container/pkg/features"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

const (
	FeatureFlagName = "no-certificate-authorities"
	FlagHelpMessage = "Disable certificate authority trust mount"

	defaultCaTrustSourceAnchorPath = "/etc/pki/ca-trust/source/anchors"
	defaultcaTrustDestinationPath  = "/etc/pki/ca-trust/source/anchors"
)

type config struct {
	Enabled    bool   `mapstructure:"enabled"`
	SourcePath string `mapstructure:"source_anchors"`
}

func newConfigWithDefaults() *config {
	config := config{}
	config.Enabled = true
	config.SourcePath = defaultCaTrustSourceAnchorPath
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
		log.Debugf("Certificate authorities disabled via config")
	}
	if viper.IsSet(FeatureFlagName) {
		log.Debugf("Certificate authorities disabled via flag")
	}
	return f.config.Enabled && !viper.IsSet(FeatureFlagName)
}

func (f *Feature) ExitOnError() bool {
	return false
}

func (f *Feature) Configure() error {
	cfg := newConfigWithDefaults()

	if !viper.IsSet("features.certificate_authorities") {
		f.config = cfg
		return nil
	}

	f.userHasConfig = true
	err := viper.UnmarshalKey("features.certificate_authorities", &cfg)
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

	sourcePath := f.config.SourcePath
	_, err := f.afs.Stat(sourcePath)
	if err != nil {
		return opts, fmt.Errorf("error: problem reading CA Anchors: %v: %w", sourcePath, err)
	}

	opts.AddVolumeMount(engine.VolumeMount{
		Source:       sourcePath,
		Destination:  defaultcaTrustDestinationPath,
		MountOptions: "ro",
	})

	return opts, nil
}

func (f *Feature) HandleError(err error) {
	if f.userHasConfig {
		log.Warnf("Error initializing certificate authorities functionality: %v", err)
	}
	log.Debugf("Error initializing certificate authorities functionality: %v", err)
}

func init() {
	f := Feature{
		afs: &afero.Afero{Fs: afero.NewOsFs()},
	}
	features.Register("certificateAuthorities", &f)
}
