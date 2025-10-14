package osdctl

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOsdctl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Osdctl Suite")
}
