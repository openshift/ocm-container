package features

import (
	"errors"
	"fmt"

	"github.com/openshift/ocm-container/pkg/engine"
	log "github.com/sirupsen/logrus"
)

type Feature interface {
	Configure() error
	Initialize() (OptionSet, error)
	Enabled() bool
	HandleError(error)
	ExitOnError() bool
}

var features map[string]Feature

type OptionSet struct {
	Mounts []engine.VolumeMount
	Envs   []engine.EnvVar
}

func (o *OptionSet) AddVolumeMount(mount ...engine.VolumeMount) {
	o.Mounts = append(o.Mounts, mount...)
}

func (o *OptionSet) AddEnv(env ...engine.EnvVar) {
	o.Envs = append(o.Envs, env...)
}

func (o *OptionSet) AddEnvKey(key string) {
	o.Envs = append(o.Envs, engine.EnvVar{Key: key})
}

func (o *OptionSet) AddEnvKeyVal(key string, val string) {

	o.Envs = append(o.Envs, engine.EnvVar{Key: key, Value: val})
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
	log.Debugf("initializing all features")
	for featureName, f := range features {
		log.Debugf("configuring feature - %s", featureName)
		err := f.Configure()
		if err != nil {
			log.Warnf("error configuring feature %s - skipping - %v", featureName, err)
			continue
		}
		if !f.Enabled() {
			log.Infof("%s - feature not enabled", featureName)
			continue
		}
		log.Debugf("feature %s configuration complete", featureName)
		log.Debugf("initializing feature - %s", featureName)
		opts, err := f.Initialize()
		if err != nil {
			f.HandleError(err)
			if f.ExitOnError() {
				terminalErrors = errors.Join(terminalErrors, err)
			}
		}
		allOptions.AddVolumeMount(opts.Mounts...)
		allOptions.AddEnv(opts.Envs...)
		log.Debugf("feature %s initialization complete", featureName)
	}
	return allOptions, terminalErrors
}

func Reset() {
	features = map[string]Feature{}
}
