package ocmcontainer

import (
	"fmt"
	"path/filepath"

	"github.com/openshift/ocm-container/pkg/engine"
)

type ocmContainer struct {
	engine *engine.Engine
}

func (o *ocmContainer) start() error {
	return nil
}

func (o *ocmContainer) exec() error {
	return nil
}

func (o *ocmContainer) copy(source, destination string) error {
	s := filepath.Clean(source)
	d := filepath.Clean(destination)

	args := fmt.Sprintf("%s:%s", s, d)

	o.engine.Copy("cp", args)

	return nil
}

func (o *ocmContainer) open() error {
	return nil
}
