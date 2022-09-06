package cli

import (
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/homedir"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// debug flag indicating that cli should output debug logs
	debug bool
	// config file location
	configFile string
	// dry run flag
	dryRun bool
	// name of the executable
	exeName = filepath.Base(os.Args[0])

	// version of package, substitute via ldflags
	ver string
)

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
		hostname := viper.GetString(apiHostFlag)
		scheme := viper.GetString(schemeFlag)
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
func MakeRootCmd(httpClient *http.Client, k8sclient kubernetes.Interface) (*cobra.Command, error) {
	cobra.OnInitialize(initViperConfigs)

	// Use executable name as the command name
	rootCmd := &cobra.Command{
		Use: exeName,
	}

	// register basic flags
	rootCmd.PersistentFlags().String(apiHostFlag, client.DefaultHost, "KuberLogic API server address")
	err := viper.BindPFlag(apiHostFlag, rootCmd.PersistentFlags().Lookup(apiHostFlag))
	if err != nil {
		return nil, err
	}
	rootCmd.PersistentFlags().String(schemeFlag, client.DefaultSchemes[0], fmt.Sprintf("KuberLogic API server scheme: %v", client.DefaultSchemes))
	err = viper.BindPFlag(schemeFlag, rootCmd.PersistentFlags().Lookup(schemeFlag))
	if err != nil {
		return nil, err
	}
	rootCmd.PersistentFlags().String(tokenFlag, "8ZTjsD3t2Q3Yq-C4-hoahcFn", "Specify KuberLogic API server authentication token. The authentication token is used for authentication to KuberLogic API every time an API request is made. Like passwords, Authentication tokens should remain a secret.")
	err = viper.BindPFlag(tokenFlag, rootCmd.PersistentFlags().Lookup(tokenFlag))
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
	rootCmd.PersistentFlags().Var(&formatResponse, formatFlag, "Format response value: json, yaml or string. (default: string)")

	// register security flags
	// add all operation groups
	rootCmd.AddCommand(
		makeServiceCmd(makeClientClosure(httpClient)),
		makeBackupCmd(makeClientClosure(httpClient)),
		makeRestoreCmd(makeClientClosure(httpClient)),

		makeInstallCmd(k8sclient),
		makeDiagCmd(),
		makeVersionCmd(k8sclient),
		makeInfoCmd(k8sclient, makeClientClosure(httpClient)),
	)

	// add cobra completion
	rootCmd.AddCommand(makeGenCompletionCmd())

	return rootCmd, nil
}

// initViperConfigs initialize viper config using config file in '$HOME/.config/<cli name>/config.<json|yaml...>'
// currently hostname, scheme and auth tokens can be specified in this config file.
func initViperConfigs() {
	if configFile == "" {
		// use default config file
		configFile = path.Join(homedir.HomeDir(), ".config", "kuberlogic", "config.yaml")
	}
	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err != nil {
		logDebugf("Error: loading config file: %v", err)
		return
	}
	logDebugf("Using config file: %v", viper.ConfigFileUsed())
}

func makeServiceCmd(apiClientFunc func() (*client.ServiceAPI, error)) *cobra.Command {
	operationGroupServiceCmd := &cobra.Command{
		Use:   "service",
		Short: `Service related operations`,
	}
	operationGroupServiceCmd.AddCommand(
		makeServiceAddCmd(apiClientFunc),
		makeServiceEditCmd(apiClientFunc),
		makeServiceDeleteCmd(apiClientFunc),
		makeServiceListCmd(apiClientFunc),
		makeServiceBackupCmd(apiClientFunc),
		makeServiceCredentialsUpdateCmd(apiClientFunc),
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
