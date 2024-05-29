package version

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/openshift/ocm-container/pkg/utils"
	"github.com/spf13/cobra"
)

type versionResponse struct {
	Commit  string `json:"commit"`
	Version string `json:"version"`
	Latest  string `json:"latest"`
}

// VersionCmd represents the version command
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version",
	Long:  `Display the version of ocm-container version`,
	RunE:  version,
}

func version(cmd *cobra.Command, args []string) error {
	gitCommit := "unknown"

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				gitCommit = setting.Value
				break
			}
		}
	}

	latest, _ := utils.GetLatestVersion() // let's ignore this error, just in case we have no internet access
	ver, err := json.MarshalIndent(&versionResponse{
		Commit:  gitCommit,
		Version: utils.Version,
		Latest:  strings.TrimPrefix(latest, "v"),
	}, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(ver))
	return nil
}
