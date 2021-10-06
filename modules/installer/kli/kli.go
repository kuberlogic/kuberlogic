package kli

import (
	"github.com/kuberlogic/kuberlogic/modules/installer/cfg"
	helm_installer "github.com/kuberlogic/kuberlogic/modules/installer/kli/helm-installer"
	logger "github.com/kuberlogic/kuberlogic/modules/installer/log"
)

type KuberlogicInstaller interface {
	Install(args []string) error
	Upgrade(args []string) error
	Uninstall(args []string) error
	Status(args []string) error
	Exit(err error)
}

func NewInstaller(config *cfg.Config, log logger.Logger) (KuberlogicInstaller, error) {
	helm, err := helm_installer.New(config, log)
	if err != nil {
		return nil, err
	}
	return helm, nil
}
