/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/openshift/ocm-container/cmd/version"
	"github.com/openshift/ocm-container/pkg/features/registrar"
	"github.com/openshift/ocm-container/pkg/log"
	"github.com/openshift/ocm-container/pkg/ocm"
	"github.com/openshift/ocm-container/pkg/ocmcontainer"
	"github.com/openshift/ocm-container/pkg/subprocess"
)

const (
	ocmcManagedNameEnv = "IO_OPENSHIFT_MANAGED_NAME"

	defaultImage = "quay.io/redhat-services-prod/openshift/ocm-container:latest"
)

var errInContainer = errors.New("already running inside ocm-container; turtles all the way down")

var vols []string
var envs []string
var execArgs []string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "ocm-container",
	Example: `
ocm-container [flags]  # opens a blank container without any cluster variables set, but logged into OCM
ocm-container [flags] -- [command]			# execute a command in the container without logging into a cluster
ocm-container --cluster-id CLUSTER_ID [flags]		# log into a cluster
ocm-container --cluster-id CLUSTER_ID [flags] -- [command]	# execute a command inside the container after logging into a cluster
`,
	Short: "Launch an OCM container",
	Long: `Launches a container with the OCM environment 
and other Red Hat SRE tools`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		e := os.Getenv(ocmcManagedNameEnv)
		if e == programName {
			return errInContainer
		}

		// if we pass the `--version` flag, just run the version command
		// and exit early.
		if versionExitEarly {
			return version.VersionCmd.RunE(cmd, args)
		}

		err := checkFlags(cmd)
		if err != nil {
			return err
		}

		// From here on out errors are application errors, not flag or argument errors
		// Don't print the help message if we get an error returned
		cmd.SilenceUsage = true

		err = log.InitializeLogger()
		if err != nil {
			return err
		}

		// Append any volumes passed in as flags to the volumes slice from the config
		viper.Set("vols", vols)

		o, err := ocmcontainer.New(
			cmd,
			execArgs,
		)
		if err != nil {
			return err
		}

		err = o.Start(false)
		if err != nil {
			return err
		}

		err = o.ExecPostRunBlockingCmds()
		if err != nil {
			return err
		}

		// Start non-blocking commands, but expect no output
		o.ExecPostRunNonBlockingCmds()

		err = o.Run()
		if err != nil {
			return err
		}

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		ocm.CloseClient()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		if execErr, ok := err.(*subprocess.ExecErr); ok {
			os.Exit(execErr.ExitErr.ExitCode())
		}
		os.Exit(1)
	}
}

// We don't want to allow any arguments to the ocm-container command, as they all should
// be flags, however we do want any arguments that are coming after a `--` to be passed
// to the container as the command to run, so we split on `--` and reset the args to let
// cobra's NoArgs parameter handle the argument parsing and pass the execArgs directly.
func splitArgs(args []string) ([]string, []string) {
	if len(args) == 0 || len(args) == 1 {
		return nil, nil
	}

	args = args[1:]

	for i, arg := range args {
		if arg == "--" {
			return args[:i], args[i+1:]
		}
	}
	return args, nil
}

func init() {
	cobra.OnInitialize(initConfig)
	var cobraArgs []string

	rootCmd.SetHelpTemplate(helpTemplate)

	cobraArgs, execArgs = splitArgs(os.Args)
	rootCmd.SetArgs(cobraArgs)

	// Persistent flags available to subcommands; see flags.go
	for _, f := range persistentFlags {
		switch f.flagType {
		case "bool":
			ptr := f.pointer.(*bool)
			if f.shorthand != "" {
				rootCmd.PersistentFlags().BoolVarP(ptr, f.name, f.shorthand, f.BoolValue(), f.HelpString())
			} else {
				rootCmd.PersistentFlags().BoolVar(ptr, f.name, f.BoolValue(), f.HelpString())
			}
		case "string":
			ptr := f.pointer.(*string)
			if f.shorthand != "" {
				rootCmd.PersistentFlags().StringVarP(ptr, f.name, f.shorthand, f.StringValue(), f.HelpString())
			} else {
				rootCmd.PersistentFlags().StringVar(ptr, f.name, f.StringValue(), f.HelpString())
			}
		}
	}

	// Standard flags; see flags.go
	for _, f := range standardFlags {
		switch f.flagType {
		case "bool":
			if f.shorthand != "" {
				rootCmd.Flags().BoolP(f.name, f.shorthand, f.BoolValue(), f.HelpString())
			} else {
				rootCmd.Flags().Bool(f.name, f.BoolValue(), f.HelpString())
			}
		case "string":
			if f.shorthand != "" {
				rootCmd.Flags().StringP(f.name, f.shorthand, f.StringValue(), f.HelpString())
			} else {
				rootCmd.Flags().String(f.name, f.StringValue(), f.HelpString())
			}
		}

		if f.hidden {
			_ = rootCmd.Flags().MarkHidden(f.name)
		}
	}

	for _, flag := range registrar.FeatureFlags() {
		rootCmd.Flags().Bool(flag.Name, false, strings.ToLower(flag.HelpMsg))
		// All feature flags are marked as hidden by default.
		rootCmd.Flags().MarkHidden(flag.Name)
	}

	rootCmd.Flags().StringArrayVarP(&vols, "volume", "v", []string{}, "Additional bind mounts to pass into the container. This flag does NOT overwrite what's in the config but appends to it")
	rootCmd.Flags().StringArrayVarP(&envs, "environment", "e", []string{}, "Additional environment variables to pass into the container. This flag does NOT overwrite what's in the config but appends to it")

	// Register sub-commands
	rootCmd.AddCommand(version.VersionCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".ocm-container" (without extension).
		viper.AddConfigPath(home + "/" + programName)
		viper.SetConfigType("yaml")
		viper.SetConfigName(programName)
	}

	viper.SetEnvPrefix(programPrefix)

	// Set viper defaults
	viper.SetDefault("engine", "podman")
	viper.SetDefault("image", defaultImage)

	// read in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		// TODO: Prompt to run the config command
		fmt.Fprintf(os.Stderr, "Error reading config file: %s\n", err)
	}
}
