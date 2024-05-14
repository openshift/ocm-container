package backplane

import (
	"os"

	"github.com/openshift/ocm-container/pkg/engine"
)

const (
	backplaneConfig          = ".config/backplane/config.json"
	backplaneConfigDest      = "/root/.config/backplane/config.json"
	backplaneConfigMountOpts = "rw"
)

var osStat = os.Stat

type Config struct {
	Env    map[string]string
	Mounts []engine.VolumeMount
}

func New(home string) (*Config, error) {
	c := &Config{}

	b := os.Getenv("BACKPLANE_CONFIG")
	if b == "" {
		b = home + "/" + backplaneConfig
	}

	_, err := osStat(b)
	if err != nil {
		return c, err
	}

	c.Env = make(map[string]string)
	c.Env["BACKPLANE_CONFIG"] = backplaneConfigDest // This will ALWAYS be the same inside the container, since we mount it

	c.Mounts = append(c.Mounts, engine.VolumeMount{
		Source:       b,
		Destination:  backplaneConfigDest,
		MountOptions: backplaneConfigMountOpts,
	})

	return c, nil
}
