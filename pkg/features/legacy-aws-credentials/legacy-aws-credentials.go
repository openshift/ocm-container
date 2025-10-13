package legacyawscredentials

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
	FeatureFlagName = "no-legacy-aws"
	FlagHelpMessage = "Disable legacy AWS credentials mounting"

	awsCredentials = ".aws/credentials"
	awsConfig      = ".aws/config"
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
		log.Debugf("Legacy AWS credentials disabled via config")
	}
	if viper.IsSet(FeatureFlagName) {
		log.Debugf("Legacy AWS credentials disabled via flag")
	}
	return f.config.Enabled && !viper.IsSet(FeatureFlagName)
}

func (f *Feature) ExitOnError() bool {
	return false
}

func (f *Feature) Configure() error {
	cfg := newConfigWithDefaults()

	if !viper.IsSet("features.legacy_aws_credentials") {
		f.config = cfg
		return nil
	}

	f.userHasConfig = true
	err := viper.UnmarshalKey("features.legacy_aws_credentials", &cfg)
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

	home := os.Getenv("HOME")
	if home == "" {
		return opts, fmt.Errorf("environment variable $HOME is not set")
	}

	for _, file := range []string{awsCredentials, awsConfig} {
		filePath := home + "/" + file
		_, err := f.afs.Stat(filePath)
		if err != nil {
			log.Infof("warning: problem reading AWS file: %v: %v", filePath, err)
		} else {
			opts.AddVolumeMount(engine.VolumeMount{
				Source:       filePath,
				Destination:  "/root/" + file,
				MountOptions: "ro",
			})
		}
	}

	return opts, nil
}

func (f *Feature) HandleError(err error) {
	if f.userHasConfig {
		log.Warnf("Error initializing legacy AWS credentials functionality: %v", err)
	}
	log.Debugf("Error initializing legacy AWS credentials functionality: %v", err)
}

func init() {
	f := Feature{
		afs: &afero.Afero{Fs: afero.NewOsFs()},
	}
	features.Register("legacyAwsCredentials", &f)
}
