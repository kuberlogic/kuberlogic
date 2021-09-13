package cmd

import (
	"github.com/spf13/cobra"
)

// newStatusCmd returns the status command
func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "shows a status of installed Kuberlogic release",
		Run: func(cmd *cobra.Command, args []string) {
			kuberlogicInstaller.Exit(kuberlogicInstaller.Status(args))
		},
	}
}
