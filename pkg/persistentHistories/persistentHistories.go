package persistentHistories

import (
	"os"
	"path/filepath"

	"github.com/openshift/ocm-container/pkg/deprecation"
	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/openshift/ocm-container/pkg/ocm"
)

const (
	histFile     = "/root/per-cluster/.bash_history"
	destDir      = "/root/per-cluster"
	sourceSubDir = "/.config/ocm-container/per-cluster-persistent/"
)

type Config struct {
	Env    map[string]string
	Mounts []engine.VolumeMount
}

func New(home, cluster string) (*Config, error) {
	var err error

	config := &Config{}

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

	config.Env["HISTFILE"] = histFile

	return config, nil
}

func DeprecatedConfig() bool {
	env := os.Getenv("PERSISTENT_CLUSTER_HISTORIES")
	if env != "" {
		deprecation.Print(
			"PERSISTENT_CLUSTER_HISTORIES",
			"enable_persistent_histories")
		return true
	}
	return false
}
