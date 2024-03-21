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
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/openshift/ocm-container/pkg/ocmcontainer"
)

const (
	programName = "ocm-container"
)

var (
	requiredFlags = map[string][]string{
		"ocm-container": {"engine"},
		"build":         {"engine"},
	}
	cfgFile string
	cfg     map[string]interface{}
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ocm-container",
	Short: "Launch an OCM container",
	Long: `Launches a container with the OCM environment 
and other Red Hat SRE tools set up.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		err := checkFlags(cmd)
		if err != nil {
			return err
		}

		verbose := func(verbose, debug bool) bool {
			return verbose || debug
		}(viper.GetBool("verbose"), viper.GetBool("debug"))

		engine := viper.GetString("engine")
		dryRun := viper.GetBool("dry-run")

		o, err := ocmcontainer.New(
			cmd,
			args,
			engine,
			dryRun,
			verbose,
		)
		if err != nil {
			return err
		}

		err = o.StartAndAttach()
		if err != nil {
			return err
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	configFileDefault := fmt.Sprintf("%s/.config/%s/%s.yaml", os.Getenv("HOME"), programName, programName)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", configFileDefault, "config file")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Local flags for ocm-container
	rootCmd.Flags().Bool("dry-run", false, "parses arguments and environment and prints the command that would be executed, but does not execute it.")
	rootCmd.Flags().Bool("verbose", false, "verbose output")
	rootCmd.Flags().Bool("debug", false, "debug output (deprecated: use --verbose. This will be removed in a future release.)")

	supportedEngines := fmt.Sprintf("container engine to use (%s)", strings.Join(engine.SupportedEngines, ", "))
	rootCmd.Flags().String("engine", "", supportedEngines)

	rootCmd.Flags().BoolP("disable-console-port", "d", false, "disable the console port mapping (Linux-only; console port Will not work with MacOS)")
	rootCmd.Flags().BoolP("no-personalizations", "n", true, "disable personalizations file ")
	rootCmd.Flags().StringP("exec", "e", "", "execute a command in a running container (deprecated: append '-- <command>' to the end of the command to execute. This will be removed in a future release.)")
	rootCmd.Flags().StringP("launch-opts", "o", "", "additional launch options for the container")
	rootCmd.Flags().StringP("image", "i", "ocm-container", "Sets the image name to sue")
	rootCmd.Flags().StringP("tag", "t", "latest", "Sets the image tag to use")
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

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	} else {
		// TODO: Prompt to run the config command
		fmt.Fprintf(os.Stderr, "Error reading config file: %s\n", err)
	}

	err := viper.Unmarshal(&cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshalling config file: %s\n", err)
	}
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
