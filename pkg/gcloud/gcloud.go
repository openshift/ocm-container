package gcloud

import "github.com/openshift/ocm-container/pkg/engine"

const (
	gcloudConfigDir = ".config/gcloud"
)

type Config struct {
	Mount engine.VolumeMount
}

func New(home string) (*Config, error) {

	config := &Config{
		Mount: engine.VolumeMount{
			Source:       home + "/" + gcloudConfigDir,
			Destination:  "/root/" + gcloudConfigDir,
			MountOptions: "ro",
		},
	}

	return config, nil
}
