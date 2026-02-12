package gcloud

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGcloud(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gcloud Suite")
}
