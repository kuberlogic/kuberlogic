package backuprestore_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBackuprestore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Backuprestore Suite")
}
