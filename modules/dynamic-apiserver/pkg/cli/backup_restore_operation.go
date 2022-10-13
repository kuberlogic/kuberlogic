package cli

import (
	"fmt"

	client2 "github.com/go-openapi/runtime/client"
	"github.com/spf13/viper"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client/restore"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"

	"github.com/spf13/cobra"
)

// makeBackupRestoreCmd returns a cmd to handle operation backupRestore
func makeBackupRestoreCmd(apiClientFunc func() (*client.ServiceAPI, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "backupRestore",
		Short:   `Creates a restore request for a service`,
		Aliases: []string{"restore"},
		RunE:    runBackupRestore(apiClientFunc),
	}

	_ = cmd.PersistentFlags().String(backupIdFlag, "", "Required. Backup ID")
	cmd.MarkFlagRequired(backupIdFlag)

	return cmd
}

// runBackupRestore uses cmd flags to call endpoint api
func runBackupRestore(apiClientFunc func() (*client.ServiceAPI, error)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		apiClient, err := apiClientFunc()
		if err != nil {
			return err
		}

		// retrieve flag values from cmd and fill params
		params := restore.NewRestoreAddParams()
		res := models.Restore{}

		if value, err := getString(cmd, backupIdFlag); err != nil {
			return err
		} else if value != nil {
			res.BackupID = *value
		}

		var formatResponse format
		if value, err := getString(cmd, formatFlag); err != nil {
			return err
		} else if value != nil {
			formatResponse = format(*value)
		}

		params.RestoreItem = &res
		if dryRun {
			logDebugf("Params: %+v", params.RestoreItem)
			logDebugf("dry-run flag specified. Skip sending request.")
			return nil
		}

		// make request and then print result
		response, err := apiClient.Restore.RestoreAdd(params,
			client2.APIKeyAuth("X-Token", "header", viper.GetString(tokenFlag)))
		if err != nil {
			return humanizeError(err)
		}
		payload := response.GetPayload()
		if isDefaultPrintFormat(formatResponse) {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "A request '%s' to restore backup '%s' successfully created\n", payload.ID, payload.BackupID)
			return err
		} else {
			return printResult(cmd, formatResponse, payload)
		}
	}
}
