package additionalclusterenvs

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAdditionalClusterEnvs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AdditionalClusterEnvs Suite")
}
