package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/openshift/ocm-container/pkg/deprecation"
	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// NOTE: FUTURE OPTIONS SHOULD NOT CONFLICT WITH PODMAN/DOCKER FLAGS
// TO ALLOW FOR PASSING IN CONTAINER-SPECIFIC OPTIONS WHEN NECESSARY
// AND TO AVOID CONFUSION

const (
	programName   = "ocm-container"
	programPrefix = "OCMC"
)

// requiredFlags maps the required flags for a given subcommand
var (
	requiredFlags = map[string][]string{
		"ocm-container": {"engine"},
		"build":         {"engine"},
	}
)

type cliFlag struct {
	pointer        interface{}
	name           string
	shorthand      string
	flagType       string
	value          string
	helpMsg        string
	deprecationMsg string
}

func (f cliFlag) StringValue() string {
	return f.value
}

// BoolValue returns the boolean value of the flag parsed from the string value
func (f cliFlag) BoolValue() bool {
	b, _ := strconv.ParseBool(f.value)
	return b
}

func (f cliFlag) HelpString() string {
	return strings.ToLower(f.helpMsg + f.deprecationMsg)
}

// persistentFlags are the list of flags that are available to all commands
// other than the configFile flag, handled separately

var (
	cfgFile string
	debug   bool
	dryRun  bool
	verbose bool
)

var configFileDefault = fmt.Sprintf("%s/.config/%s/%s.yaml", os.Getenv("HOME"), programName, programName)

var persistentFlags = []cliFlag{
	{
		pointer:  &cfgFile,
		name:     "config",
		flagType: "string",
		value:    configFileDefault,
		helpMsg:  "config file to use",
	},
	{
		pointer:        &debug,
		name:           "debug",
		shorthand:      "x",
		flagType:       "bool",
		value:          "false",
		helpMsg:        "Enable debug output",
		deprecationMsg: deprecation.ShortMessage("--debug", "--verbose"),
	},
	{
		pointer:  &dryRun,
		name:     "dry-run",
		flagType: "bool",
		value:    "false",
		helpMsg:  "Parses arguments and environment and prints the command that would be executed, but does not execute it.",
	},
	{
		pointer:   &verbose,
		name:      "verbose",
		shorthand: "v",
		flagType:  "bool",
		value:     "false",
		helpMsg:   "Enable verbose output",
	},
}

var standardFlags = []cliFlag{
	{
		name:     "cluster-id",
		flagType: "string",
		helpMsg:  "Optional cluster ID to log into on launch",
	},
	{
		name:     "engine",
		flagType: "string",
		helpMsg:  fmt.Sprintf("Container engine to use (%s)", strings.Join(engine.SupportedEngines, ", ")),
	},
	{
		name:     "headless",
		flagType: "string",
		helpMsg:  "Run the container in the background (no console)",
	},
	{
		name:     "launch-opts",
		flagType: "string",
		helpMsg:  "Additional container engine launch options for the container",
	},
	{
		name:           "exec",
		shorthand:      "e", // -e is already in use by podman; this should be migrated to -E or replaced by the container CMD
		flagType:       "string",
		helpMsg:        "Execute a command in a running container",
		deprecationMsg: deprecation.ShortMessage("--exec", "append '-- [command]'. See --help for examples"),
	},
	{
		name:     "entrypoint",
		flagType: "string",
		helpMsg:  "Overwrite the default ENTRYPOINT of the image",
	},
	{

		name:      "registry",
		shorthand: "R",
		flagType:  "string",
		value:     "quay.io",
		helpMsg:   "Sets the image registry to use",
	},
	{
		name:      "repository",
		shorthand: "O",
		flagType:  "string",
		value:     "app-sre",
		helpMsg:   "Sets the image repository organization to use",
	},
	{
		name:      "image",
		shorthand: "I",
		flagType:  "string",
		value:     "ocm-container",
		helpMsg:   "Sets the image name to use",
	},
	{
		name:      "tag",
		shorthand: "t", // -t is already in use by podman; this should be migrated to -T
		flagType:  "string",
		value:     "latest",
		helpMsg:   "Sets the image tag to use",
	},
}

// disableFeatureFlags is a list of feature flags can be used to disable features of the container,
// These features can be disabled via CLI flags, or Viper environment variables or configuration file.

var disabledFeaturesHelpMessage = `
Flags with the prefix '--no-' can be used to disable features of the container,
particularly those that mount volumes or specify extra environment variables.

In addition to CLI flags, these features can be disabled via Viper environment variables or configuration file. 

For example: 

	'no-aws' can be set to 'true' in the configuration file, or 
    'OCMC_NO_AWS' can be set to 'TRUE' in the environment to disable AWS CLI mounts and environment.
`

var disableFeatureFlags = []cliFlag{
	{
		name:    "no-aws",
		helpMsg: "Disable AWS CLI mounts and environment",
	},
	{
		name:    "no-certificate-authorities",
		helpMsg: "Disable mounting the host's certificate authorities",
	},
	{
		name:           "disable-console-port",
		helpMsg:        "Disable the console port mapping",
		deprecationMsg: deprecation.ShortMessage("--disable-console-port", "--no-console-port"),
	},
	{
		name:    "no-console-port",
		helpMsg: "Disable the console port mapping",
	},
	{
		name:    "no-gcp",
		helpMsg: "Disable Google Cloud (GCP) mounts and environment",
	},
	{
		name:    "no-jira",
		helpMsg: "Disable JIRA CLI mounts and environment",
	},
	{
		name:    "no-ops-utils",
		helpMsg: "Disable ops-utils mounts and environment",
	},
	{
		name:    "no-osdctl",
		helpMsg: "Disable OSDCTL mounts and environment",
	},
	{
		name:    "no-pagerduty",
		helpMsg: "Disable PagerDuty CLI mounts and environment",
	},
	{
		name:    "no-persistent-histories",
		helpMsg: "Disable persistent histories file mounts and environment",
	},
	{
		name:    "no-personalizations",
		helpMsg: "Disable personalizations file mounts and environment",
	},
	{
		name:    "no-scratch",
		helpMsg: "Disable scratch directory mounts and environment",
	},
}

// checkFlags looks up the required flags for the given cobra.Command,
// checks if they are set in viper, and returns an error if they are not.
func checkFlags(cmd *cobra.Command) error {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		viper.BindPFlag(f.Name, f)
	})

	val, ok := requiredFlags[cmd.Use]

	if ok {
		for _, flag := range val {
			if !viper.IsSet(flag) {
				return fmt.Errorf("required flag %s not set", flag)
			}
		}
	}
	return nil
}
