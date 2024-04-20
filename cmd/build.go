/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/openshift/ocm-container/pkg/subprocess"
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

		out, err := subprocess.RunLive(c)
		if err != nil {
			return err
		}

		fmt.Println(out)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	supportedEngines := fmt.Sprintf("container engine to use (%s)", strings.Join(engine.SupportedEngines, ", "))
	buildCmd.Flags().String("engine", "", supportedEngines)

	buildCmd.Flags().StringP("build-args", "b", "", "container engine build arguments (eg: --no-cache)")
	buildCmd.Flags().StringP("image", "i", programName, "container image name")
	buildCmd.Flags().StringP("tag", "t", "latest", "container image tag")
}
