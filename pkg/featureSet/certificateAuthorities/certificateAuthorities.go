package certificateAuthorities

import (
	"fmt"
	"os"

	"github.com/openshift/ocm-container/pkg/engine"
)

const (
	caTrustSourceAnchorPath = "/etc/pki/ca-trust/source/anchors"
)

type Config struct {
	Mounts []engine.VolumeMount
}

func New(hostCaPath string) (*Config, error) {
	var err error

	config := &Config{}

	var source string
	if hostCaPath == "" {
		_, err = os.Stat(caTrustSourceAnchorPath)
		source = caTrustSourceAnchorPath
	} else {
		_, err = os.Stat(hostCaPath)
		source = hostCaPath
	}
	if err != nil {
		return config, fmt.Errorf("error: problem reading CA Anchors: %v: %v, err", source, err)
	}

	config.Mounts = append(config.Mounts, engine.VolumeMount{
		Source:       source,
		Destination:  caTrustSourceAnchorPath,
		MountOptions: "ro",
	})

	return config, nil
}
