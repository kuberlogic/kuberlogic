package cmd

import (
	"github.com/spf13/cobra"
)

// newUninstallCmd returns the delete command
func newUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "uninstall",
		Short:     "uninstall a Kuberlogic release",
		ValidArgs: []string{"force"},
		Run: func(cmd *cobra.Command, args []string) {
			kuberlogicInstaller.Exit(kuberlogicInstaller.Uninstall(args))
		},
	}
}
