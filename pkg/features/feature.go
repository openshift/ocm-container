package features

import (
	"errors"
	"fmt"

	"github.com/openshift/ocm-container/pkg/engine"
)

type Feature interface {
	Initialize() (OptionSet, error)
	Enabled() bool
	HandleError(error)
	ExitOnError() bool
}

var features map[string]Feature

type OptionSet struct {
	Mounts []engine.VolumeMount
}

func (o *OptionSet) AddVolumeMount(mount ...engine.VolumeMount) {
	o.Mounts = append(o.Mounts, mount...)
}

func NewOptionSet() OptionSet {
	o := OptionSet{}
	o.Mounts = []engine.VolumeMount{}

	return o
}

func Register(name string, feature Feature) error {
	if features == nil {
		features = map[string]Feature{}
	}

	if _, ok := features[name]; ok {
		return fmt.Errorf("feature %s already registered", name)
	}

	features[name] = feature
	return nil
}

func Initialize() (OptionSet, error) {
	var terminalErrors error

	allOptions := NewOptionSet()
	for _, f := range features {
		if !f.Enabled() {
			continue
		}
		opts, err := f.Initialize()
		if err != nil {
			f.HandleError(err)
			if f.ExitOnError() {
				terminalErrors = errors.Join(terminalErrors, err)
			}
		}
		allOptions.AddVolumeMount(opts.Mounts...)
	}
	return allOptions, terminalErrors
}
