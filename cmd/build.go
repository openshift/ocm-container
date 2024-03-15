/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		var makeCmd string = "make"
		var makeTarget string = "build"
		var makeArgs strings.Builder

		var e string = fmt.Sprintf("CONTAINER_ENGINE=%s", containerEngine)

		makeArgs.WriteString(strings.Join(args, " ") + makeTarget)

		c := exec.Command(makeCmd, makeArgs.String())
		c.Env = append(c.Env, e)

		image, err := cmd.Flags().GetString("image")
		if err != nil {
			fmt.Print(err)
			return
		}
		c.Env = append(c.Env, "IMAGE_NAME="+image)

		tag, err := cmd.Flags().GetString("tag")
		if err != nil {
			fmt.Print(err)
			return
		}
		c.Env = append(c.Env, "IMAGE_TAG="+tag)

		buildArgs, err := cmd.Flags().GetString("build-args")
		if err != nil {
			fmt.Print(err)
			return
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
			fmt.Print(err)
			return
		}

		// stderr is the pipe for err output
		var stdErr io.ReadCloser
		stdErr, err = c.StderrPipe()
		if err != nil {
			fmt.Print(err)
			return
		}

		err = c.Start()
		if err != nil {
			fmt.Print(err)
			return
		}

		var out []byte
		out, err = io.ReadAll(stdOut)
		if err != nil {
			fmt.Print(err)
			return
		}

		var errOut []byte
		errOut, err = io.ReadAll(stdErr)
		if err != nil {
			fmt.Print(err)
			return
		}

		if errOut != nil {
			fmt.Fprint(os.Stderr, string(errOut))
			return
		}

		fmt.Print(string(out))
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

	buildCmd.Flags().StringP("build-args", "b", "", "Container engine build arguments (eg: --no-cache)")
	buildCmd.Flags().StringP("image", "i", programName, "Container image name")
	buildCmd.Flags().StringP("tag", "t", "latest", "Container image tag")
}
