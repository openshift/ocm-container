package osdctl

import "github.com/openshift/ocm-container/pkg/engine"

const (
	osdctlConfigDir = ".config/osdctl"
)

type Config struct {
	Mounts []engine.VolumeMount
}

func New(home string) (*Config, error) {

	config := &Config{}

	config.Mounts = append(config.Mounts, engine.VolumeMount{
		Source:       home + "/" + osdctlConfigDir,
		Destination:  "/root/" + osdctlConfigDir,
		MountOptions: "ro",
	})

	return config, nil
}
