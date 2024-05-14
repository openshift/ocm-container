package persistentImages

import (
	"os"
	"path/filepath"

	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/openshift/ocm-container/pkg/ocm"
)

const (
	destDir   = "/var/lib/containers/storage/"
	sourceDir = "/.cache/ocm-container/images/"
)

type Config struct {
	Env    map[string]string
	Mounts []engine.VolumeMount
}

func New(home string) (*Config, error) {
	var err error

	config := &Config{}

	ocmClient, err := ocm.NewClient()
	if err != nil {
		return config, err
	}
	defer ocmClient.Close()

	mount := filepath.Join(home, sourceDir)
	err = os.MkdirAll(mount, os.ModePerm)
	if err != nil {
		return config, err
	}

	config.Mounts = append(config.Mounts, engine.VolumeMount{
		Source:       mount,
		Destination:  destDir,
		MountOptions: "rw",
	})

	return config, nil
}
