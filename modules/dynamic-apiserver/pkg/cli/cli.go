package cli

import (
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// debug flag indicating that cli should output debug logs
var debug bool

// config file location
var configFile string

// dry run flag
var dryRun bool

// name of the executable
var exeName string = filepath.Base(os.Args[0])

// logDebugf writes debug log to stdout
func logDebugf(format string, v ...interface{}) {
	if !debug {
		return
	}
	log.Printf(format, v...)
}

// makeClient constructs a client object
func makeClientClosure(httpClient *http.Client) func() (*client.ServiceAPI, error) {
	return func() (*client.ServiceAPI, error) {
		hostname := viper.GetString("hostname")
		scheme := viper.GetString("scheme")
		logDebugf("hostname %s, scheme %s", hostname, scheme)

		r := httptransport.NewWithClient(hostname, client.DefaultBasePath, []string{scheme}, httpClient)
		r.SetDebug(debug)
		// set custom producer and consumer to use the default ones

		r.Consumers["application/json"] = runtime.JSONConsumer()
		r.Producers["application/json"] = runtime.JSONProducer()

		appCli := client.New(r, strfmt.Default)
		logDebugf("Server url: %v://%v", scheme, hostname)
		return appCli, nil
	}
}

// MakeRootCmd returns the root cmd
func MakeRootCmd(httpClient *http.Client) (*cobra.Command, error) {
	cobra.OnInitialize(initViperConfigs)

	// Use executable name as the command name
	rootCmd := &cobra.Command{
		Use: exeName,
	}

	// register basic flags
	rootCmd.PersistentFlags().String("hostname", client.DefaultHost, "hostname of the service")
	err := viper.BindPFlag("hostname", rootCmd.PersistentFlags().Lookup("hostname"))
	if err != nil {
		return nil, err
	}
	rootCmd.PersistentFlags().String("scheme", client.DefaultSchemes[0], fmt.Sprintf("Choose from: %v", client.DefaultSchemes))
	err = viper.BindPFlag("scheme", rootCmd.PersistentFlags().Lookup("scheme"))
	if err != nil {
		return nil, err
	}
	rootCmd.PersistentFlags().String("token", "", "authentication apiserver token")
	err = viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	if err != nil {
		return nil, err
	}

	// configure debug flag
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "output debug logs")
	// configure config location
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file path")
	// configure dry run flag
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "do not send the request to server")

	var formatResponse format
	rootCmd.PersistentFlags().Var(&formatResponse, format_flag, "Format response value: json, yaml or string. (default: string)")

	// register security flags
	// add all operation groups
	rootCmd.AddCommand(
		makeServiceCmd(makeClientClosure(httpClient)),
		makeBackupCmd(makeClientClosure(httpClient)),
		makeRestoreCmd(makeClientClosure(httpClient)),
	)

	// add cobra completion
	rootCmd.AddCommand(makeGenCompletionCmd())

	return rootCmd, nil
}

// initViperConfigs initialize viper config using config file in '$HOME/.config/<cli name>/config.<json|yaml...>'
// currently hostname, scheme and auth tokens can be specified in this config file.
func initViperConfigs() {
	if configFile != "" {
		// use user specified config file location
		viper.SetConfigFile(configFile)
	} else {
		// look for default config
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(path.Join(home, ".config", exeName))
		viper.SetConfigName("config")
	}

	if err := viper.ReadInConfig(); err != nil {
		logDebugf("Error: loading config file: %v", err)
		return
	}
	logDebugf("Using config file: %v", viper.ConfigFileUsed())
}

func makeServiceCmd(apiClientFunc func() (*client.ServiceAPI, error)) *cobra.Command {
	operationGroupServiceCmd := &cobra.Command{
		Use:  "service",
		Long: `Service related operations`,
	}
	operationGroupServiceCmd.AddCommand(
		makeServiceAddCmd(apiClientFunc),
		makeServiceEditCmd(apiClientFunc),
		makeServiceDeleteCmd(apiClientFunc),
		makeServiceListCmd(apiClientFunc),
		makeServiceBackupCmd(apiClientFunc),
	)

	return operationGroupServiceCmd
}

func makeBackupCmd(apiClientFunc func() (*client.ServiceAPI, error)) *cobra.Command {
	operationGroupBackupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Backups related operations",
	}

	operationGroupBackupCmd.AddCommand(
		makeBackupDeleteCmd(apiClientFunc),
		makeBackupListCmd(apiClientFunc),
		makeBackupRestoreCmd(apiClientFunc),
	)

	return operationGroupBackupCmd
}

func makeRestoreCmd(apiClientFunc func() (*client.ServiceAPI, error)) *cobra.Command {
	operationGroupRestoreCmd := &cobra.Command{
		Use:   "restore",
		Short: "Restores related operations",
	}

	operationGroupRestoreCmd.AddCommand(
		makeRestoreDeleteCmd(apiClientFunc),
		makeRestoreListCmd(apiClientFunc),
	)
	return operationGroupRestoreCmd
}
