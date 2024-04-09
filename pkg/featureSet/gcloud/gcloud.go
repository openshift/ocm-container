package gcloud

import (
	"fmt"
	"os"

	"github.com/openshift/ocm-container/pkg/engine"
)

const (
	gcloudConfigDir = ".config/gcloud"
)

type Config struct {
	Mounts []engine.VolumeMount
}

func New(home string) (*Config, error) {

	config := &Config{}

	_, err := os.Stat(home + "/" + gcloudConfigDir)
	if err != nil {
		fmt.Printf("warning: problem reading gcloud config dir: %v: %v\n", home+"/"+gcloudConfigDir, err)
	} else {
		config.Mounts = append(config.Mounts, engine.VolumeMount{
			Source:       home + "/" + gcloudConfigDir,
			Destination:  "/root/" + gcloudConfigDir,
			MountOptions: "ro",
		})
	}

	return config, nil
}
