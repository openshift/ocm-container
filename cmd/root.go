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

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	easy "github.com/t-tomalak/logrus-easy-formatter"

	"github.com/openshift/ocm-container/pkg/ocmcontainer"
	"github.com/openshift/ocm-container/pkg/subprocess"
)

const (
	ocmcManagedNameEnv = "IO_OPENSHIFT_MANAGED_NAME"
)

var errInContainer = errors.New("already running inside ocm-container; turtles all the way down")

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "ocm-container",
	Example: `
ocm-container [flags]
ocm-container [flags] -- - [command]			# execute a command in the container without logging into a cluster
ocm-container --cluster-id CLUSTER_ID [flags]		# log into a cluster
ocm-container --cluster-id CLUSTER_ID [flags] -- [command]	# execute a command inside the container after logging into a cluster

ocm-container [flags] cluster_id		# log into a cluster; deprecated: use '--cluster-id CLUSTER_ID'
ocm-container [flags] -- cluster_id [command]	# execute a command inside the container after logging into a cluster; deprecated: use '--cluster-id CLUSTER_ID'
`,
	Short: "Launch an OCM container",
	Long: `Launches a container with the OCM environment 
and other Red Hat SRE tools`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		e := os.Getenv(ocmcManagedNameEnv)
		if e == programName {
			return errInContainer
		}

		err := checkFlags(cmd)
		if err != nil {
			return err
		}

		log.SetFormatter(&easy.Formatter{
			LogFormat: "[%lvl%]: %msg%\n",
		})
		log.SetLevel(setLogLevel(viper.GetBool("verbose"), viper.GetBool("debug")))

		// From here on out errors are application errors, not flag or argument errors
		// Don't print the help message if we get an error returned
		cmd.SilenceUsage = true

		o, err := ocmcontainer.New(
			cmd,
			args,
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

		if !viper.GetBool("--headless") {
			o.Attach()
		}

		return nil
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

func init() {
	cobra.OnInitialize(initConfig)

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
	}

	// Disable features list; see flags.go
	for _, flag := range disableFeatureFlags {
		rootCmd.Flags().Bool(flag.name, false, strings.ToLower(flag.helpMsg+flag.deprecationMsg))
	}

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
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	} else {
		// TODO: Prompt to run the config command
		fmt.Fprintf(os.Stderr, "Error reading config file: %s\n", err)
	}
}

func setLogLevel(verbose, debug bool) log.Level {
	if debug {
		return log.DebugLevel
	}
	if verbose {
		return log.InfoLevel
	}
	return log.WarnLevel
}
