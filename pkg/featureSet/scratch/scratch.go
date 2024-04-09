package scratch

import (
	"fmt"
	"os"

	"github.com/openshift/ocm-container/pkg/engine"
)

const (
	destinationDir = "/root/scratch"
)

type Config struct {
	Mounts []engine.VolumeMount
}

func New(source string, rw bool) (*Config, error) {
	var err error

	config := &Config{}

	_, err = os.Stat(source)
	if err != nil {
		return config, fmt.Errorf("error: problem reading scratch directory: %v: %v,", source, err)
	}

	config.Mounts = append(config.Mounts, engine.VolumeMount{
		Source:       source,
		Destination:  destinationDir,
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
