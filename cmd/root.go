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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/openshift/ocm-container/pkg/ocmcontainer"
)

const (
	programName = "ocm-container"
)

var cfgFile string
var verbose bool
var debug bool
var containerEngine string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ocm-container",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	Args: rootArgs,
	Run: func(cmd *cobra.Command, args []string) {

		v := func(verbose, debug bool) bool {
			return verbose || debug
		}(verbose, debug)

		o, err := ocmcontainer.New(cmd, args, containerEngine, v)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		err = o.StartAndAttach()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

func rootArgs(cmd *cobra.Command, args []string) error {
	return nil
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

	configFileDefault := fmt.Sprintf("%s/.config/%s/.%s.yaml", os.Getenv("HOME"), programName, programName)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", configFileDefault, "config file")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "debug output (Deprecated: use --verbose. This will be removed in a future release.)")
	rootCmd.PersistentFlags().StringVar(&containerEngine, "engine", "", "container engine to use (podman, docker)")
	rootCmd.MarkPersistentFlagRequired("engine")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Local flags for ocm-container
	rootCmd.Flags().BoolP("disable-console-port", "d", false, "Disable the console port mapping (Linux-only; console port Will not work with MacOS)")
	rootCmd.Flags().BoolP("no-personalizations", "n", true, "Disable personalizations file ")
	rootCmd.Flags().StringP("exec", "e", "", "Execute a command in a running container (Deprecated: append '-- <command>' to the end of the command to execute. This will be removed in a future release.)")
	rootCmd.Flags().StringP("launch-opts", "o", "", "Additional launch options for the container")
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
		viper.SetConfigName("." + programName)
		viper.AutomaticEnv()
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	} else {
		// TODO: Prompt to run the config command
		fmt.Fprintf(os.Stderr, "Error reading config file: %s\n", err)
	}
}
