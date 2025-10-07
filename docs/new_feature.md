# Adding a New Feature

Adding a new feature is simple.

Copy the following scaffolding to a new folder in the `pkg/features` directory:

```go
package myfeature

// Define any defaults here as constants
const (
    FeatureFlagName = "no-myFeature"
    FlagHelpMessage = "Disables myFeature functionality"

    defaultMyConfigVar = "someString"
)

// Any internal config needed for the setup of the feature
type config struct {
	MyConfigVar string `mapstructure:"my_config_var"`
}

// This is where we want to set all of our config defaults. If
// the user doesn't explicitly NEED to set something, set it
// here for them and allow them to overwrite it.
func newConfigWithDefaults() *config {
	config := config{}
    config.MyConfigVar = defaultMyConfigVar
	return &config
}

// Validate is where any custom configuration validation logic
// lives. This is where you need to validate your user's input
func (cfg *config) validate() error {
	return nil
}

type Feature struct{}

// Enabled is where we determine whether or not the feature
// is explicitly enabled if opt-in or disabled if opt-out.
func (f *Feature) Enabled() bool {
	return true
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

// Initialize is the feature that we use to create the OptionSet
// for a given feature. An OptionSet is how the ocm-container
// program knows what options to pass into the container create
// command in order for the individual feature to work properly
func (f *Feature) Initialize() (features.OptionSet, error) {
	opts := features.NewOptionSet()

	cfg := newConfigWithDefaults()

	viper.UnmarshalKey("features.myFeature", &cfg)
	err := cfg.validate()
	if err != nil {
		return opts, err
	}

    ... more logic here ...

	opts.AddVolumeMount(engine.VolumeMount{
		Source:       outOfContainerSource,
		Destination:  inContainerDestination,
	})

	return opts, nil
}

// If initialize fails, how should we handle the error? This
// allows you to customize what log level to use or how to
// clean up anything you need to.
func (f *Feature) HandleError(err error) {
	log.Warnf("Error initializing PagerDuty functionality: %v", err)
}

func init() {
    f := Feature{}
    features.Register("myFeature", &f)
}
```

Once the above scaffolding is filled out for your feature, then add the feature import to the [feature registrar](/pkg/features/registrar/registrar.go) and reference the `FeatureFlagName` and `FlagHelpMessage` consts from the featureFlags list. This does two things - 1. Forces you to think about how someone might want to disable this via the command line and 2. allows the rest of the feature to be initialized and registered via the init() function inside the feature.

The feature flag to disable the feature SHOULD opt to use a `--no-myFeature` convention. In certain cases it might be more gramatically correct to `--disable-myFeature` but lets opt for brevity unless it really makes sense. This can be decided on a case-by-case basis.

## Configuration

Using the `viper.UnmarshalKey("features.myFeature")` convention above, and by applying the `mapstructure:"myKey"` annotation in the config struct, we are able to structure our `ocm-container` `config.yaml` file with the following structure:

```yaml
engine: podman
...
features:
  myFeature:
    myKey: myValue
```

This allows each function to define it's own feature set, and even allows overlapping keys between functions, since they're nested in their various config structs.

However, the only convention that we will enforce is to use camelCase for names in the config file as well as to reserve the key "enabled" to be a boolean value for each feature. We should strive for consistency so that if our users want to disable features they should be able to relatively quickly assume that it would an entry of `enabled: false` for that feature configuration.
