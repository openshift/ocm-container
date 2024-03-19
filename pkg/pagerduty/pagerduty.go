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
	Token string
	Mount engine.VolumeMount
}

func New(home string) (*Config, error) {
	var err error

	config := &Config{
		Token: fmt.Sprintf("%s/%s", home, pagerDutyTokenFile),
	}

	_, err = os.Stat(config.Token)
	if err != nil {
		return config, fmt.Errorf("error: problem reading PagerDuty token file: %v: %v, err", config.Token, err)
	}

	config.Mount = engine.VolumeMount{
		Source:       config.Token,
		Destination:  pagerDutyTokenDest,
		MountOptions: "ro",
	}

	return config, nil
}
