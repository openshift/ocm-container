package features

import "github.com/openshift/ocm-container/pkg/engine"

type Feature interface {
	New() (*OptionSet, error)
}

type OptionSet struct {
	Mounts []engine.VolumeMount
}

func NewOptionSet() OptionSet {
	o := OptionSet{}
	o.Mounts = []engine.VolumeMount{}

	return o
}

func (o *OptionSet) AddVolumeMount(mount engine.VolumeMount) {
	o.Mounts = append(o.Mounts, mount)
}
