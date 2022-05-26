package compose_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCompose(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Compose Suite")
}
