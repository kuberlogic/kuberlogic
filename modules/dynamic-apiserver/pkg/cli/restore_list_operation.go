package cli

import (
	client2 "github.com/go-openapi/runtime/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client/restore"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
	"strconv"

	"github.com/spf13/cobra"
)

// makeRestoreListCmd returns a cmd to handle operation restoreList
func makeRestoreListCmd(apiClientFunc func() (*client.ServiceAPI, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "restoreList",
		Short:   `List of restore objects`,
		Aliases: []string{"list"},
		RunE:    runRestoreList(apiClientFunc),
	}

	_ = cmd.PersistentFlags().String(serviceIdFlag, "", "service id to filter by")
	return cmd
}

// runRestoreList uses cmd flags to call endpoint api
func runRestoreList(apiClientFunc func() (*client.ServiceAPI, error)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		apiClient, err := apiClientFunc()
		if err != nil {
			return err
		}

		params := restore.NewRestoreListParams()
		if value, err := getString(cmd, serviceIdFlag); err != nil {
			return err
		} else if value != nil {
			params.ServiceID = value
		}

		var formatResponse format
		if value, err := getString(cmd, formatFlag); err != nil {
			return err
		} else if value != nil {
			formatResponse = format(*value)
		}

		if dryRun {
			logDebugf("dry-run flag specified. Skip sending request.")
			return nil
		}

		response, err := apiClient.Restore.RestoreList(params,
			client2.APIKeyAuth("X-Token", "header", viper.GetString(tokenFlag)))
		if err != nil {
			return humanizeError(err)
		}

		payload := response.GetPayload()
		if isDefaultPrintFormat(formatResponse) {
			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.SetHeader([]string{"№", "ID", "Backup ID", "Created", "Status"})
			table.SetBorder(false)
			for i, item := range payload {
				table.Append([]string{
					strconv.Itoa(i), item.ID, item.BackupID, item.CreatedAt.String(), item.Status})
			}
			table.Render()
		} else {
			return printResult(cmd, formatResponse, payload)
		}
		return nil
	}
}
