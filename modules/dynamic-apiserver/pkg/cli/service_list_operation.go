package cli

import (
	"strconv"

	client2 "github.com/go-openapi/runtime/client"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client/service"

	"github.com/spf13/cobra"
)

// makeServiceListCmd returns a cmd to handle operation serviceList
func makeServiceListCmd(apiClientFunc func() (*client.ServiceAPI, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serviceList",
		Short:   `List of service objects`,
		Aliases: []string{"list"},
		RunE:    runServiceList(apiClientFunc),
	}

	_ = cmd.PersistentFlags().String(subscriptionId, "", "Subscription id to filter by")

	return cmd
}

// runServiceList uses cmd flags to call endpoint api
func runServiceList(apiClientFunc func() (*client.ServiceAPI, error)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		apiClient, err := apiClientFunc()
		if err != nil {
			return err
		}

		params := service.NewServiceListParams()

		var formatResponse format
		if value, err := getString(cmd, "format"); err != nil {
			return err
		} else if value != nil {
			formatResponse = format(*value)
		}

		if dryRun {
			logDebugf("dry-run flag specified. Skip sending request.")
			return nil
		}

		response, err := apiClient.Service.ServiceList(params,
			client2.APIKeyAuth("X-Token", "header", viper.GetString(tokenFlag)))
		if err != nil {
			return humanizeError(err)
		}

		payload := response.GetPayload()
		if isDefaultPrintFormat(formatResponse) {
			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.SetHeader([]string{"№", "ID", "Subscription ID", "Type", "Replica", "Version", "Backup Schedule", "Status", "Endpoint"})
			table.SetBorder(false)
			for i, item := range payload {
				table.Append([]string{
					strconv.Itoa(i), *item.ID, item.Subscription, *item.Type, strconv.Itoa(int(*item.Replicas)),
					item.Version, item.BackupSchedule, item.Status, item.Endpoint,
				})
			}
			table.Render()
		} else {
			return printResult(cmd, formatResponse, payload)
		}
		return nil
	}
}
