package backplane

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBackplane(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Backplane Suite")
}
