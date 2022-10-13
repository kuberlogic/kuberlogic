package cli

import (
	"fmt"

	client2 "github.com/go-openapi/runtime/client"
	"github.com/spf13/viper"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"

	"github.com/spf13/cobra"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client/restore"
)

// makeRestoreDeleteCmd returns a cmd to handle operation restoreDelete
func makeRestoreDeleteCmd(apiClientFunc func() (*client.ServiceAPI, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "restoreDelete",
		Short:   `Deletes a restore object`,
		Aliases: []string{"delete"},
		RunE:    runRestoreDelete(apiClientFunc),
	}

	_ = cmd.PersistentFlags().String(idFlag, "", "Required. Restore id")
	cmd.MarkFlagRequired(idFlag)

	return cmd
}

// runRestoreDelete uses cmd flags to call endpoint api
func runRestoreDelete(apiClientFunc func() (*client.ServiceAPI, error)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		apiClient, err := apiClientFunc()
		if err != nil {
			return err
		}

		// retrieve flag values from cmd and fill params
		params := restore.NewRestoreDeleteParams()

		if value, err := getString(cmd, idFlag); err != nil {
			return err
		} else if value != nil {
			params.RestoreID = *value
		}

		var formatResponse format
		if value, err := getString(cmd, formatFlag); err != nil {
			return err
		} else if value != nil {
			formatResponse = format(*value)
		}

		if dryRun {
			logDebugf("Params: %+v", params.RestoreID)
			logDebugf("dry-run flag specified. Skip sending request.")
			return nil
		}

		// make request and then print result
		response, err := apiClient.Restore.RestoreDelete(params,
			client2.APIKeyAuth("X-Token", "header", viper.GetString(tokenFlag)))
		if err != nil {
			return humanizeError(err)
		}
		if isDefaultPrintFormat(formatResponse) {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "Restore '%s' successfully deleted\n", params.RestoreID)
			return err
		} else {
			return printResult(cmd, formatResponse, response)
		}
	}
}
