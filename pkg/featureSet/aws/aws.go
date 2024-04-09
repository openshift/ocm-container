package aws

import (
	"fmt"
	"os"

	"github.com/openshift/ocm-container/pkg/engine"
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
			fmt.Printf("warning: problem reading AWS file: %v: %v\n", home+"/"+file, err)
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
