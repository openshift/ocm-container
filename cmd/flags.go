package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/openshift/ocm-container/pkg/deprecation"
	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/openshift/ocm-container/pkg/ocm"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// NOTE: FUTURE OPTIONS SHOULD NOT CONFLICT WITH PODMAN/DOCKER FLAGS
// TO ALLOW FOR PASSING IN CONTAINER-SPECIFIC OPTIONS WHEN NECESSARY
// AND TO AVOID CONFUSION
// The exception to this rule is when a flag is essentially passed
// directly into the container engine's run command.
// For example, we have a `-v "/path/to/file:/container/path/file"`
// flag that (while validated) will essentially map directly to the
// same flag for the run command. These are okay, because they add
// a better user experience

const (
	programName   = "ocm-container"
	programPrefix = "OCMC"
)

// requiredFlags maps the required flags for a given subcommand
var (
	requiredFlags = map[string][]string{
		"ocm-container": {"engine", "ocm-url"},
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

	// hidden, when set to true, will not show the flag
	// as part of the `--help` output
	hidden bool
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
		pointer:   &debug,
		name:      "debug",
		shorthand: "x",
		flagType:  "bool",
		value:     "false",
		helpMsg:   "Enable debug output",
	},
	{
		pointer:  &dryRun,
		name:     "dry-run",
		flagType: "bool",
		value:    "false",
		helpMsg:  "Parses arguments and environment and prints the command that would be executed, but does not execute it.",
	},
}

var standardFlags = []cliFlag{
	{
		name:      "cluster-id",
		flagType:  "string",
		shorthand: "C",
		helpMsg:   "Optional cluster ID to log into on launch",
	},
	{
		name:     "engine",
		flagType: "string",
		helpMsg:  fmt.Sprintf("Container engine to use (%s)", strings.Join(engine.SupportedEngines, ", ")),
	},
	{
		name:     "ocm-url",
		flagType: "string",
		value:    "prod",
		helpMsg:  fmt.Sprintf("OCM Environment (%s)", strings.Join(ocm.SupportedUrls, ", ")),
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
		name:     "pull",
		flagType: "string",
		value:    "always",
		helpMsg:  fmt.Sprintf("Pull image policy (%s)", strings.Join(engine.SupportedPullImagePolicies, ", ")),
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
	{
		name:     "publish-all-ports",
		flagType: "bool",
		value:    "false",
		helpMsg:  "Publishes all defined ports to all interfaces. Equivalent of `--publish-all`",
		hidden:   true,
	},
}

// disableFeatureFlags is a list of feature flags can be used to disable features of the container,
// These features can be disabled via CLI flags, or Viper environment variables or configuration file.

var disableFeatureFlags = []cliFlag{
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
		name:    "no-ops-utils",
		helpMsg: "Disable ops-utils mounts and environment",
	},
	{
		name:    "no-osdctl",
		helpMsg: "Disable OSDCTL mounts and environment",
	},
	{
		name:    "no-persistent-histories",
		helpMsg: "Disable persistent histories file mounts and environment",
	},
	{
		name:    "no-persistent-images",
		helpMsg: "Disable local container storage cache mount",
	},
	{
		name:    "no-personalizations",
		helpMsg: "Disable personalizations file mounts and environment",
	},
}

// checkFlags looks up the required flags for the given cobra.Command,
// checks if they are set in viper, and returns an error if they are not.
func checkFlags(cmd *cobra.Command) error {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		err := viper.BindPFlag(f.Name, f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error binding flag %s: %v\n", f.Name, err)
			os.Exit(1)
		}
	})

	val, ok := requiredFlags[cmd.Use]

	if ok {
		for _, flag := range val {
			if (!viper.IsSet(flag)) && (viper.GetString(flag) == "") {
				return fmt.Errorf("required flag %s not set", flag)
			}
		}
	}
	return nil
}
