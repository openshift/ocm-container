package aws

import (
	"fmt"
	"os"

	"github.com/openshift/ocm-container/pkg/engine"
	log "github.com/sirupsen/logrus"
)

const (
	awsCredentials = ".aws/credentials"
	awsConfig      = ".aws/config"
)

type Config struct {
	Mounts []engine.VolumeMount
}

func New(home string) (*Config, error) {
	var err error

	config := &Config{}

	for _, file := range []string{awsCredentials, awsConfig} {
		_, err = os.Stat(home + "/" + file)
		if err != nil {
			log.Warn(fmt.Sprintf("warning: problem reading AWS file: %v: %v\n", home+"/"+file, err))
		} else {
			config.Mounts = append(config.Mounts, engine.VolumeMount{
				Source:       home + "/" + file,
				Destination:  "/root/" + file,
				MountOptions: "ro",
			})
		}
	}

	return config, nil
}
