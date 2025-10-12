package jira

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

// Define any defaults here as constants
const (
	FeatureFlagName = "no-jira"
	FlagHelpMessage = "Disables jira functionality"

	jiraEnvTokenKey           = "JIRA_API_TOKEN"
	jiraAuthTypeKey           = "JIRA_AUTH_TYPE"
	defaultAuthType           = "bearer"
	defaultConfigFileLocation = ".config/.jira/.config.yml"

	jiraConfigFileDest = "/root/.config/.jira/.config.yml"
)

// Any internal config needed for the setup of the feature
type config struct {
	Enabled   bool   `mapstructure:"enabled"`
	FilePath  string `mapstructure:"config_file"`
	MountOpts string `mapstructure:"config_mount"`
}

// This is where we want to set all of our config defaults. If
// the user doesn't explicitly NEED to set something, set it
// here for them and allow them to overwrite it.
func newConfigWithDefaults() *config {
	config := config{}
	config.Enabled = true
	config.FilePath = defaultConfigFileLocation
	config.MountOpts = "ro"
	return &config
}

// Validate is where any custom configuration validation logic
// lives. This is where you need to validate your user's input
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
	afs    *afero.Afero

	// If the user provided a configuration, set that here
	// in case we want to handle initialization errors
	// differently because of that
	userHasConfig bool
}

// Enabled is where we determine whether or not the feature
// is explicitly enabled if opt-in or disabled if opt-out.
func (f *Feature) Enabled() bool {
	if !f.config.Enabled {
		log.Debugf("JIRA disabled via config")
	}
	if viper.IsSet(FeatureFlagName) {
		log.Debugf("JIRA disabled via flag")
	}
	return f.config.Enabled && !viper.IsSet(FeatureFlagName)
}

// If this feature is required for the functionality of
// ocm-container OR if a configuration error will be
// catastrophic to our user's experience, set this to true.
// Otherwise, if we lose a convenience function but we should
// still allow the user to use the container, then set false.
// In almost all cases, this should be set to false.
func (f *Feature) ExitOnError() bool {
	return false
}

// We want to self-contain the configuration functionality separate
// from the initialization so that we can read in the enabled config
func (f *Feature) Configure() error {
	cfg := newConfigWithDefaults()

	if !viper.IsSet("features.jira") {
		f.config = cfg
		return nil
	}
	f.userHasConfig = true
	err := viper.UnmarshalKey("features.jira", &cfg)
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

// Initialize is the feature that we use to create the OptionSet
// for a given feature. An OptionSet is how the ocm-container
// program knows what options to pass into the container create
// command in order for the individual feature to work properly
func (f *Feature) Initialize() (features.OptionSet, error) {
	log.Debug("Initializing JIRA Options")
	opts := features.NewOptionSet()

	if os.Getenv(jiraEnvTokenKey) != "" {
		// env key is set, let's handle without checking for token file
		opts.AddEnvKey(jiraEnvTokenKey)
		if os.Getenv(jiraAuthTypeKey) != "" {
			opts.AddEnvKey(jiraAuthTypeKey)
		} else {
			opts.AddEnvKeyVal(jiraAuthTypeKey, defaultAuthType)
		}
	}

	jiraConfigFile, err := f.statConfigFileLocations()
	if err != nil {
		return opts, err
	}
	opts.AddVolumeMount(engine.VolumeMount{
		Source:      jiraConfigFile,
		Destination: jiraConfigFileDest,
	})

	return opts, nil
}

// If initialize fails, how should we handle the error? This
// allows you to customize what log level to use or how to
// clean up anything you need to.
func (f *Feature) HandleError(err error) {
	// example how we want to handle feature intilization
	// errors differently based on whether or not the user
	// provided a configuration. In this case, if the user
	// provided config themselves we want to explicitly warn
	// them of the error, but if they didn't set a config and
	// the default functionality isn't working, there's no
	// need to inform them because they didn't set it up.
	if f.userHasConfig {
		log.Warnf("Error initializing JIRA functionality: %v", err)
	}
	log.Debugf("Error initializing JIRA functionality: %v", err)
}

// check for config file locations in the following order:
// absolute path -> $HOME/(path)
// return error if not found after all have been checked
func (f *Feature) statConfigFileLocations() (string, error) {
	filepath := f.config.FilePath
	if filepath == "" {
		return "", fmt.Errorf("no filepath provided")
	}
	errorPaths := []string{}
	_, err := f.afs.Stat(filepath)
	if err == nil {
		log.Debugf("using %s for JIRA config", filepath)
		return filepath, nil
	}
	errorPaths = append(errorPaths, filepath)

	path := os.Getenv("HOME") + "/" + filepath
	_, err = f.afs.Stat(path)
	if err == nil {
		log.Debugf("using %s for JIRA config", path)
		return path, nil
	}
	errorPaths = append(errorPaths, path)

	return "", fmt.Errorf("could not find %s in any of: %s", filepath, errorPaths)
}
func init() {
	f := Feature{
		afs: &afero.Afero{Fs: afero.NewOsFs()},
	}
	features.Register("jira", &f)
}
