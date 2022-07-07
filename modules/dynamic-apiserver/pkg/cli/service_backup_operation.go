package cli

import (
	"fmt"
	client2 "github.com/go-openapi/runtime/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"
	"github.com/spf13/viper"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client/backup"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"

	"github.com/spf13/cobra"
)

// makeServiceAddCmd returns a cmd to handle operation serviceAdd
func makeServiceBackupCmd(apiClientFunc func() (*client.ServiceAPI, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serviceBackup",
		Short:   `Creates a backup request for a service`,
		Aliases: []string{"backup"},
		RunE:    runServiceBackup(apiClientFunc),
	}

	_ = cmd.PersistentFlags().String(service_id_flag, "", "service id")

	return cmd
}

// runServiceAdd uses cmd flags to call endpoint api
func runServiceBackup(apiClientFunc func() (*client.ServiceAPI, error)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		apiClient, err := apiClientFunc()
		if err != nil {
			return err
		}

		// retrieve flag values from cmd and fill params
		params := backup.NewBackupAddParams()
		bak := models.Backup{}

		if value, err := getString(cmd, service_id_flag); err != nil {
			return err
		} else if value != nil {
			bak.ServiceID = *value
		}

		var formatResponse format
		if value, err := getString(cmd, format_flag); err != nil {
			return err
		} else if value != nil {
			formatResponse = format(*value)
		}

		params.BackupItem = &bak
		if dryRun {
			logDebugf("Params: %+v", params.BackupItem)
			logDebugf("dry-run flag specified. Skip sending request.")
			return nil
		}

		// make request and then print result
		response, err := apiClient.Backup.BackupAdd(params,
			client2.APIKeyAuth("X-Token", "header", viper.GetString("token")))
		if err != nil {
			return humanizeError(err)
		}
		payload := response.GetPayload()
		if isDefaultPrintFormat(formatResponse) {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "A request for backup '%s' successfully created\n", payload.ID)
			return err
		} else {
			return printResult(cmd, formatResponse, payload)
		}

	}
}
