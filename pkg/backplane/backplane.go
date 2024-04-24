package backplane

import (
	"fmt"
	"os"

	"github.com/openshift/ocm-container/pkg/engine"
)

const (
	backplaneConfig     = ".config/backplane/config.json"
	backplaneConfigDest = "/root/.config/backplane/config.json"
)

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

	_, err := os.Stat(b)
	if err != nil {
		return c, fmt.Errorf("error: problem reading backplane config: %v: %v", b, err)
	}

	c.Env = make(map[string]string)
	c.Env["BACKPLANE_CONFIG"] = backplaneConfig
	c.Mounts = append(c.Mounts, engine.VolumeMount{
		Source:       b,
		Destination:  backplaneConfigDest,
		MountOptions: "rw",
	})

	return c, nil
}
