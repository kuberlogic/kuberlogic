package cmd

import (
	"github.com/kuberlogic/operator/modules/installer/kli"
	"github.com/spf13/cobra"
)

// newUpgradeCmd returns the upgrade command
func newUpgradeCmd(k kli.KuberlogicInstaller) *cobra.Command {
	return &cobra.Command{
		Use:       "upgrade",
		ValidArgs: []string{installAllArg, installCertManagerArg, installDepsArg, installKuberlogicArg},
		Args:      cobra.ExactValidArgs(1),
		Short:     "upgrades already installed Kuberlogic release",
		Run: func(cmd *cobra.Command, args []string) {
			k.Exit(k.Upgrade(args))
		},
	}
}
