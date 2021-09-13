package cmd

import (
	"github.com/kuberlogic/operator/modules/installer/kli"

	"github.com/spf13/cobra"
)

// newStatusCmd returns the status command
func newStatusCmd(k kli.KuberlogicInstaller) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "shows a status of installed Kuberlogic release",
		Run: func(cmd *cobra.Command, args []string) {
			k.Exit(k.Status(args))
		},
	}
}
