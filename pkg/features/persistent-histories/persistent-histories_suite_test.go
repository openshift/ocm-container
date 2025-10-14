package persistenthistories

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPersistentHistories(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PersistentHistories Suite")
}
