package gcloud

import (
	"fmt"
	"os"

	"github.com/openshift/ocm-container/pkg/engine"
	log "github.com/sirupsen/logrus"
)

const (
	gcloudConfigDir = ".config/gcloud"
	gcloudLogDir = ".config/gcloud/logs"
)

type Config struct {
	Mounts []engine.VolumeMount
}

func New(home string) (*Config, error) {

	config := &Config{}

	_, err := os.Stat(home + "/" + gcloudConfigDir)
	if err != nil {
		log.Warn(fmt.Sprintf("warning: problem reading gcloud config dir: %v: %v\n", home+"/"+gcloudConfigDir, err))
	} else {
		config.Mounts = append(config.Mounts, engine.VolumeMount{
			Source:       home + "/" + gcloudConfigDir,
			Destination:  "/root/" + gcloudConfigDir,
			MountOptions: "ro",
		})
		config.Mounts = append(config.Mounts, engine.VolumeMount{
			Source:       home + "/" + gcloudLogDir,
			Destination:  "/root/" + gcloudLogDir,
			MountOptions: "rw",
		})
	}

	return config, nil
}
