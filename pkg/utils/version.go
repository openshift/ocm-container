package utils

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	VersionAPIEndpoint     = "https://api.github.com/repos/openshift/ocm-container/releases/latest"
	VersionAddressTemplate = "https://github.com/openshift/ocm-container/releases/download/v%s/ocm-container_%s_%s_%s.tar.gz" // version, version, GOOS, GOARCH
)

var (
	// GitCommit is the short git commit hash from the environment
	// Will be set during build process via GoReleaser
	// See also: https://pkg.go.dev/cmd/link
	GitCommit string

	// Version is the tag version from the environment
	// Will be set during build process via GoReleaser
	// See also: https://pkg.go.dev/cmd/link
	Version string = "v0.0.0-unknown"
)

// IsRunningInOcmContainer returns true if the current process is running inside
// an ocm-container environment. This is determined by checking the
// IO_OPENSHIFT_MANAGED_NAME environment variable.
//
// This function can be imported by other Go programs (such as osdctl) to detect
// when they are running inside ocm-container and adjust their behavior accordingly
// (e.g., skipping browser auto-launch, adjusting network binding addresses).
//
// Example usage:
//
//	import "github.com/openshift/ocm-container/pkg/utils"
//
//	if utils.IsRunningInOcmContainer() {
//	    // Skip browser launch, show URL instead
//	    fmt.Println("Running in container - open URL in your host browser")
//	} else {
//	    // Normal behavior
//	    openBrowser(url)
//	}
func IsRunningInOcmContainer() bool {
	return os.Getenv("IO_OPENSHIFT_MANAGED_NAME") == "ocm-container"
}

// GetOcmContainerComponent returns the component variant of the ocm-container
// image if running inside one, or an empty string if not running in ocm-container.
//
// Possible return values:
//   - "micro": ocm-container-micro variant (minimal tools: ocm, ocm-backplane, oc)
//   - "minimal": ocm-container-minimal variant (adds AWS CLI, rosa, osdctl, servicelogger)
//   - "full": ocm-container variant (full toolset including omc, jira-cli, oc-nodepp, vault, and more)
//   - "": not running in ocm-container
//
// Example usage:
//
//	component := utils.GetOcmContainerComponent()
//	if component != "" {
//	    fmt.Printf("Running in ocm-container-%s\n", component)
//	}
func GetOcmContainerComponent() string {
	if !IsRunningInOcmContainer() {
		return ""
	}
	return os.Getenv("IO_OPENSHIFT_MANAGED_COMPONENT")
}

type gitHubResponse struct {
	TagName string `json:"tag_name"`
}

// getLatestVersion connects to the GitHub API and returns the latest ocm-container tag name
func GetLatestVersion() (latest string, err error) {
	client := http.Client{
		Timeout: time.Second * 10,
	}

	// TODO: if env var Github Token is set, use that for version check
	req, err := http.NewRequest(http.MethodGet, VersionAPIEndpoint, nil)
	if err != nil {
		return latest, err
	}

	res, err := client.Do(req)
	if err != nil {
		return latest, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return latest, err
	}

	githubResp := gitHubResponse{}
	err = json.Unmarshal(body, &githubResp)
	if err != nil {
		return latest, err
	}

	return githubResp.TagName, nil
}
