package cmd

import (
	"github.com/kuberlogic/operator/modules/installer/kli"
	"github.com/spf13/cobra"
)

// newUninstallCmd returns the delete command
func newUninstallCmd(k kli.KuberlogicInstaller) *cobra.Command {
	return &cobra.Command{
		Use:       "uninstall",
		Short:     "uninstall a Kuberlogic release",
		ValidArgs: []string{"force"},
		Run: func(cmd *cobra.Command, args []string) {
			k.Exit(k.Uninstall(args))
		},
	}
}
