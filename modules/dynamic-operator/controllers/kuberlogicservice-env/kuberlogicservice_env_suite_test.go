package kuberlogicservice_env_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKuberlogicserviceEnv(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KuberlogicserviceEnv Suite")
}
