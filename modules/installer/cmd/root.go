/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/installer/cfg"
	"github.com/kuberlogic/kuberlogic/modules/installer/kli"
	logger "github.com/kuberlogic/kuberlogic/modules/installer/log"
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

	version   string // version of package
	sha1ver   string // sha1 revision used to build the program
	buildTime string // when the executable was built
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
	log.Infof("version %s, build time: %s, sha1ver: %s", version, buildTime, sha1ver)
	log.Infof("Reading config from %s", cfgFile)

	if _, err = os.Stat(cfgFile); err == nil {
		// get config
		config, err = cfg.NewConfigFromFile(cfgFile, log)
		if err != nil {
			log.Fatalf("Error reading config file: %+v", err)
		}
	} else {
		config = cfg.AskConfig(log, getDefaultCfgLocation())
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
	defaultCfgLocation := getDefaultCfgLocation()
	cobra.OnInitialize(initState)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", defaultCfgLocation, fmt.Sprintf("config file (default is %s)", defaultCfgLocation))
}

func getDefaultCfgLocation() string {
	return fmt.Sprintf("%s/%s", os.Getenv("HOME"), ".kuberlogic-installer.yaml")
}
