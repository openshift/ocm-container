package deprecation

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDeprecation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Deprecation Suite")
}
