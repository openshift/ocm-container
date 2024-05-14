package persistentImages

import (
	"os"
	"path/filepath"

	"github.com/openshift/ocm-container/pkg/engine"
	log "github.com/sirupsen/logrus"
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
	mount := filepath.Join(home, sourceDir)
	log.Debug("using image cache mount: " + mount)

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
