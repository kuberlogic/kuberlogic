package cli

import (
	"fmt"

	client2 "github.com/go-openapi/runtime/client"
	"github.com/spf13/viper"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"

	"github.com/spf13/cobra"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client/backup"
)

// makeBackupDeleteCmd returns a cmd to handle operation backupDelete
func makeBackupDeleteCmd(apiClientFunc func() (*client.ServiceAPI, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "backupDelete",
		Short:   `Deletes a backup by ID`,
		Aliases: []string{"delete"},
		RunE:    runBackupDelete(apiClientFunc),
	}

	_ = cmd.PersistentFlags().String(idFlag, "", "Required. Backup ID.")
	cmd.MarkFlagRequired(idFlag)

	return cmd
}

// runBackupDelete uses cmd flags to call endpoint api
func runBackupDelete(apiClientFunc func() (*client.ServiceAPI, error)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		apiClient, err := apiClientFunc()
		if err != nil {
			return err
		}

		// retrieve flag values from cmd and fill params
		params := backup.NewBackupDeleteParams()

		if value, err := getString(cmd, idFlag); err != nil {
			return err
		} else if value != nil {
			params.BackupID = *value
		}

		var formatResponse format
		if value, err := getString(cmd, formatFlag); err != nil {
			return err
		} else if value != nil {
			formatResponse = format(*value)
		}

		if dryRun {
			logDebugf("Params: %+v", params.BackupID)
			logDebugf("dry-run flag specified. Skip sending request.")
			return nil
		}

		// make request and then print result
		response, err := apiClient.Backup.BackupDelete(params,
			client2.APIKeyAuth("X-Token", "header", viper.GetString(tokenFlag)))
		if err != nil {
			return humanizeError(err)
		}
		if isDefaultPrintFormat(formatResponse) {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "Backup '%s' successfully deleted\n", params.BackupID)
			return err
		} else {
			return printResult(cmd, formatResponse, response)
		}

	}
}
