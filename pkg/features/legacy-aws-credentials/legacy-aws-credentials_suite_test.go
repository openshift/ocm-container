package legacyawscredentials

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLegacyAwsCredentials(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LegacyAwsCredentials Suite")
}
