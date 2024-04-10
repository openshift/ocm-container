package jira

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/openshift/ocm-container/pkg/engine"
)

const (
	jiraConfigDir     = ".config/.jira"
	jiraTokenFile     = jiraConfigDir + "/token.json"
	jiraDutyTokenDest = "/root/" + jiraTokenFile
	jiraTokenEnv      = "JIRA_API_TOKEN"
	jiraAuthTypeEnv   = "JIRA_AUTH_TYPE"
)

type Config struct {
	Token  string
	Env    map[string]string
	Mounts []engine.VolumeMount
}

func New(home string, rw bool) (*Config, error) {
	var err error

	config := &Config{}

	config.Mounts = append(config.Mounts, engine.VolumeMount{
		Source:       home + "/" + jiraConfigDir,
		Destination:  "/root/" + jiraConfigDir,
		MountOptions: boolToMountOpt(rw),
	})

	config.Env = make(map[string]string)

	// Get the token and auth type from Env
	token := os.Getenv(jiraTokenEnv)
	authType := os.Getenv(jiraAuthTypeEnv)
	if token != "" && authType != "" {
		config.Env[jiraTokenEnv] = token
		config.Env[jiraAuthTypeEnv] = authType

		return config, nil
	}

	// Else we need to read the token from the file
	t := home + "/" + jiraTokenFile
	_, err = os.Stat(t)
	if err != nil {
		return config, fmt.Errorf("error: problem reading Jira token file: %v: %v, err", t, err)
	}

	f, err := os.Open(t)
	if err != nil {
		return config, fmt.Errorf("error: problem reading Jira token file: %v: %v, err", t, err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return config, fmt.Errorf("error: problem reading Jira token file: %v: %v, err", t, err)
	}

	json.Unmarshal(b, &token)
	config.Env[jiraTokenEnv] = token
	config.Env[jiraAuthTypeEnv] = "bearer"

	return config, nil
}

// If set, mount RW
// TODO: This doesn't work for SELinux-enabled systems.
// Included for feature compatibility with previous version, but should be modified for :Z or other solution
func boolToMountOpt(rw bool) string {
	if rw {
		return "rw"
	}
	return "ro"
}
