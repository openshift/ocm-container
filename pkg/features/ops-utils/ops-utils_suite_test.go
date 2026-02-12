package opsutils

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOpsUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OpsUtils Suite")
}
