package cli

import (
	"strconv"

	client2 "github.com/go-openapi/runtime/client"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client/backup"

	"github.com/spf13/cobra"
)

// makeBackupListCmd returns a cmd to handle operation backupList
func makeBackupListCmd(apiClientFunc func() (*client.ServiceAPI, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "backupList",
		Short:   `List of backup objects`,
		Aliases: []string{"list"},
		RunE:    runBackupList(apiClientFunc),
	}

	_ = cmd.PersistentFlags().String(serviceIdFlag, "", "Service id to filter by")
	return cmd
}

// runBackupList uses cmd flags to call endpoint api
func runBackupList(apiClientFunc func() (*client.ServiceAPI, error)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		apiClient, err := apiClientFunc()
		if err != nil {
			return err
		}

		params := backup.NewBackupListParams()

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

		response, err := apiClient.Backup.BackupList(params,
			client2.APIKeyAuth("X-Token", "header", viper.GetString(tokenFlag)))
		if err != nil {
			return humanizeError(err)
		}

		payload := response.GetPayload()
		if isDefaultPrintFormat(formatResponse) {
			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.SetHeader([]string{"№", "ID", "Service ID", "Created", "Status"})
			table.SetBorder(false)
			for i, item := range payload {
				table.Append([]string{
					strconv.Itoa(i), item.ID, item.ServiceID, item.CreatedAt.String(), item.Status})
			}
			table.Render()
		} else {
			return printResult(cmd, formatResponse, payload)
		}
		return nil
	}
}
