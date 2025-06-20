package osdctl

import (
	"os"

	"github.com/openshift/ocmcontainer/pkg/engine"
)

const (
	osdctlConfigDir = ".config/osdctl"
	vaultTokenFile  = ".vaulttoken"
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

	if _, err := os.Stat(home + "/" + vaultTokenFile); err == nil {
		config.Mounts = append(config.Mounts, engine.VolumeMount{
			Source:       home + "/" + vaultTokenFile,
			Destination:  "/root/" + vaultTokenFile,
			MountOptions: "rw",
		})
	}

	return config, nil
}
