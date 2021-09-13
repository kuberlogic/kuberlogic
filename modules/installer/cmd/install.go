package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

const (
	installAllArg         = "all"
	installCertManagerArg = "cert-manager"
	installDepsArg        = "dependencies"
	installKuberlogicArg  = "kuberlogic"
)

// newInstallCmd returns the "install" command
func newInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:       fmt.Sprintf("install [%s | %s | %s | %s]", installAllArg, installCertManagerArg, installDepsArg, installKuberlogicArg),
		ValidArgs: []string{installAllArg, installCertManagerArg, installDepsArg, installKuberlogicArg},
		Args:      cobra.ExactValidArgs(1),
		Short:     "installs a Kuberlogic release",
		Run: func(cmd *cobra.Command, args []string) {
			kuberlogicInstaller.Exit(kuberlogicInstaller.Install(args))
		},
	}
}
