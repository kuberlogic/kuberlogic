package cmd

import (
	"fmt"
	"github.com/kuberlogic/operator/modules/installer/cfg"
	"github.com/kuberlogic/operator/modules/installer/kli"
	logger "github.com/kuberlogic/operator/modules/installer/log"
	"github.com/spf13/cobra"
	"os"
)

var (
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kuberlogic-installer",
	Short: "CLI for installation and management of a Kuberlogic release",
	Long: `kuberlogic-installer allows to install, manage or delete a Kuberlogic release.

Read more about how to use it on https://kuberlogic.com/docs or reach out to help@kuberlogic.com
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// initialize logger
	log := logger.NewLogger()
	log.Infof("Reading config from %s", cfgFile)

	// get config
	config, err := cfg.NewConfigFromFile(cfgFile, log)
	if err != nil {
		log.Fatalf("Error reading config file: %+v", err)
	}
	log.Debugf("Config is %+v", config)

	kuberlogicInstaller, err := kli.NewInstaller(config, log)
	if err != nil {
		log.Fatalf("Error initializing installer: %+v", err)
	}
	log.Debugf("Initialized kuberlogic installer: %+v", kuberlogicInstaller)

	// add commands
	rootCmd.AddCommand(
		newUninstallCmd(kuberlogicInstaller),
		newInstallCmd(kuberlogicInstaller),
		newUpgradeCmd(kuberlogicInstaller),
		newStatusCmd(kuberlogicInstaller),
	)
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	defaultCfgLocation := fmt.Sprintf("%s/%s", os.Getenv("HOME"), ".kuberlogic-installer.yaml")
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", defaultCfgLocation, fmt.Sprintf("config file (default is %s)", defaultCfgLocation))
}
