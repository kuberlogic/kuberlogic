package cli

import (
	client2 "github.com/go-openapi/runtime/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/client/service"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
	"strconv"

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
			client2.APIKeyAuth("X-Token", "header", viper.GetString("token")))
		if err != nil {
			return humanizeError(err)
		}

		payload := response.GetPayload()
		if isDefaultPrintFormat(formatResponse) {
			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.SetHeader([]string{"№", "ID", "Type", "Replica", "Version", "Domain", "Status", "Endpoint"})
			table.SetBorder(false)
			for i, item := range payload {
				table.Append([]string{
					strconv.Itoa(i), *item.ID, *item.Type, strconv.Itoa(int(*item.Replicas)),
					item.Version, item.Domain, item.Status, item.Endpoint,
				})
			}
			table.Render()
		} else {
			return printResult(cmd, formatResponse, payload)
		}
		return nil
	}
}
