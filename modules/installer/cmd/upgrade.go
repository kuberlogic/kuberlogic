package cmd

import (
	"github.com/spf13/cobra"
)

// newUpgradeCmd returns the upgrade command
func newUpgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "upgrade",
		ValidArgs: []string{installAllArg, installCertManagerArg, installDepsArg, installKuberlogicArg},
		Args:      cobra.ExactValidArgs(1),
		Short:     "upgrades already installed Kuberlogic release",
		Run: func(cmd *cobra.Command, args []string) {
			kuberlogicInstaller.Exit(kuberlogicInstaller.Upgrade(args))
		},
	}
}
