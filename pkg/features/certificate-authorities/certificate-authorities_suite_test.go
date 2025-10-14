package certificateauthorities

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCertificateAuthorities(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CertificateAuthorities Suite")
}
