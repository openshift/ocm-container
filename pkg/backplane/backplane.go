package backplane

import (
	"fmt"
	"os"

	"github.com/openshift/ocm-container/pkg/engine"
)

const (
	backplaneConfigDir     = ".config/backplane"
	backplaneConfigDestDir = "/root/.config/backplane"
)

type Config struct {
	Env   map[string]string
	Mount engine.VolumeMount
}

func New(home string) (*Config, error) {
	c := &Config{}

	d := os.Getenv("BACKPLANE_CONFIG_DIR")
	if d == "" {
		d = home + "/" + backplaneConfigDir
	}

	var config string = "config.json"
	ocmUrl := os.Getenv("OCM_URL")
	if ocmUrl != "" && ocmUrl != "production" {
		config = fmt.Sprintf("config.%s.json", ocmUrl)
	}

	_, err := os.Stat(d + "/" + config)
	if err != nil {
		return c, fmt.Errorf("error: problem reading backplane config: %v: %v, err", d+config, err)
	}

	c.Env = make(map[string]string)
	c.Env["BACKPLANE_CONFIG"] = config
	c.Mount = engine.VolumeMount{
		Source:       d + "/" + config,
		Destination:  backplaneConfigDestDir + "/" + config,
		MountOptions: "ro",
	}

	return c, nil
}
