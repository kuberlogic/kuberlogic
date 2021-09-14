package cmd

import (
	"fmt"
	"github.com/kuberlogic/operator/modules/installer/cfg"
	"github.com/kuberlogic/operator/modules/installer/kli"
	logger "github.com/kuberlogic/operator/modules/installer/log"
	"github.com/spf13/cobra"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kuberlogic-installer",
	Short: "CLI for installation and management of a Kuberlogic release",
	Long: `kuberlogic-installer allows to install, manage or delete a Kuberlogic release.

Read more about how to use it on https://kuberlogic.com/docs or reach out to help@kuberlogic.com
`,
}

var (
	cfgFile string

	log                 logger.Logger
	config              *cfg.Config
	kuberlogicInstaller kli.KuberlogicInstaller
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// add commands
	rootCmd.AddCommand(
		newUninstallCmd(),
		newInstallCmd(),
		newUpgradeCmd(),
		newStatusCmd(),
	)
	cobra.CheckErr(rootCmd.Execute())
}

func initState() {
	var err error
	// initialize logger with debug logs by default
	log = logger.NewLogger(true)
	log.Infof("Reading config from %s", cfgFile)

	// get config
	config, err = cfg.NewConfigFromFile(cfgFile, log)
	if err != nil {
		log.Fatalf("Error reading config file: %+v", err)
	}
	log = logger.NewLogger(*config.DebugLogs)
	log.Debugf("Config is %+v", config)

	kuberlogicInstaller, err = kli.NewInstaller(config, log)
	if err != nil {
		log.Fatalf("Error initializing installer: %+v", err)
	}
	log.Debugf("Initialized kuberlogic installer: %+v", kuberlogicInstaller)

}

func init() {
	cobra.OnInitialize(initState)

	defaultCfgLocation := fmt.Sprintf("%s/%s", os.Getenv("HOME"), ".kuberlogic-installer.yaml")
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", defaultCfgLocation, fmt.Sprintf("config file (default is %s)", defaultCfgLocation))
}
