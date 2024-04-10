package pagerduty

import (
	"fmt"
	"os"

	"github.com/openshift/ocm-container/pkg/engine"
)

const (
	pagerDutyTokenFile = ".config/pagerduty-cli/config.json"
	pagerDutyTokenDest = "/root/" + pagerDutyTokenFile
)

type Config struct {
	Token  string
	Mounts []engine.VolumeMount
}

func New(home string, rw bool) (*Config, error) {
	var err error

	config := &Config{
		Token: fmt.Sprintf("%s/%s", home, pagerDutyTokenFile),
	}

	_, err = os.Stat(config.Token)
	if err != nil {
		return config, fmt.Errorf("error: problem reading PagerDuty token file: %v: %v, err", config.Token, err)
	}

	config.Mounts = append(config.Mounts, engine.VolumeMount{
		Source:       config.Token,
		Destination:  pagerDutyTokenDest,
		MountOptions: boolToMountOpt(rw),
	})

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
