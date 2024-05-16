package persistentHistories

import (
	"os"
	"path/filepath"

	"github.com/openshift/ocm-container/pkg/deprecation"
	"github.com/openshift/ocm-container/pkg/engine"
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

func New(home, clusterId string) (*Config, error) {
	var err error

	config := &Config{}

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
			"Persistent histories will be enabled by default; use --no-persistent-histories to disable")
		return true
	}
	return false
}
