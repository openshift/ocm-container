package imagecache

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestImageCache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ImageCache Suite")
}
