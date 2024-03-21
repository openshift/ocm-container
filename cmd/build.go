/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// buildCmd builds the container image for the project locally
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds the container image for the project locally",
	Long: `The build command builds the container image for the project.
This is just a shortcut for running "make build" with image, tag, and
build option arguments, and is not required to build the image.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		err := checkFlags(cmd)
		if err != nil {
			return err
		}

		verbose := func(verbose, debug bool) bool {
			return verbose || debug
		}(viper.GetBool("verbose"), viper.GetBool("debug"))

		engine := viper.GetString("engine")

		var makeCmd string = "make"
		var makeTarget string = "build"
		var makeArgs strings.Builder

		var e string = fmt.Sprintf("CONTAINER_ENGINE=%s", engine)

		makeArgs.WriteString(strings.Join(args, " ") + makeTarget)

		c := exec.Command(makeCmd, makeArgs.String())
		c.Env = append(c.Env, e)

		image, err := cmd.Flags().GetString("image")
		if err != nil {
			return err
		}
		c.Env = append(c.Env, "IMAGE_NAME="+image)

		tag, err := cmd.Flags().GetString("tag")
		if err != nil {
			return err
		}
		c.Env = append(c.Env, "IMAGE_TAG="+tag)

		buildArgs, err := cmd.Flags().GetString("build-args")
		if err != nil {
			return err
		}
		if buildArgs != "" {
			c.Env = append(c.Env, "BUILD_ARGS="+buildArgs)
		}

		if verbose {
			fmt.Printf("Running: %s %s\n", makeCmd, makeArgs.String())
			fmt.Printf("Environment: %s\n", c.Env)
		}

		// stdOut is the pipe for command output
		// TODO: How do we stream this live?
		var stdOut io.ReadCloser
		stdOut, err = c.StdoutPipe()
		if err != nil {
			return err
		}

		// stderr is the pipe for err output
		var stdErr io.ReadCloser
		stdErr, err = c.StderrPipe()
		if err != nil {
			return err
		}

		err = c.Start()
		if err != nil {
			return err
		}

		var out []byte
		out, err = io.ReadAll(stdOut)
		if err != nil {
			return err
		}

		var errOut []byte
		errOut, err = io.ReadAll(stdErr)
		if err != nil {
			return err
		}

		if errOut != nil {
			return err
		}

		fmt.Print(string(out))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// buildCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// buildCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	buildCmd.Flags().Bool("verbose", false, "verbose output")

	supportedEngines := fmt.Sprintf("container engine to use (%s)", strings.Join(engine.SupportedEngines, ", "))
	buildCmd.Flags().String("engine", "", supportedEngines)

	buildCmd.Flags().StringP("build-args", "b", "", "container engine build arguments (eg: --no-cache)")
	buildCmd.Flags().StringP("image", "i", programName, "container image name")
	buildCmd.Flags().StringP("tag", "t", "latest", "container image tag")
}
