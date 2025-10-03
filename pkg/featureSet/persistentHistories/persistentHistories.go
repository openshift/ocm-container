package persistentHistories

import (
	"os"
	"path/filepath"

	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/openshift/ocm-container/pkg/ocm"
)

const (
	histFile     = ".bash_history"
	destDir      = "/root/.per-cluster"
	sourceSubDir = "/.config/ocm-container/per-cluster-persistent/"
)

type Config struct {
	Env    map[string]string
	Mounts []engine.VolumeMount
}

func New(home, cluster string) (*Config, error) {
	var err error

	config := &Config{}
	config.Env = make(map[string]string)

	ocmClient, err := ocm.NewClient()
	if err != nil {
		return config, err
	}
	defer ocmClient.Close()

	clusterId, err := ocm.GetClusterId(ocmClient, cluster)
	if err != nil {
		return config, err
	}

	mount := filepath.Join(home, sourceSubDir, clusterId)
	err = os.MkdirAll(mount, os.ModePerm)
	if err != nil {
		return config, err
	}

	config.Mounts = append(config.Mounts, engine.VolumeMount{
		Source:       mount,
		Destination:  destDir,
		MountOptions: "rw",
	})

	config.Env["HISTFILE"] = destDir + "/" + histFile

	return config, nil
}
