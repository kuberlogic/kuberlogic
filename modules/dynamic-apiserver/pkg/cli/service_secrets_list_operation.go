package cli

import (
	openapiClient "github.com/go-openapi/runtime/client"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client/service"

	"github.com/spf13/cobra"
)

// makeServiceAddCmd returns a cmd to handle operation serviceAdd
func makeServiceSecretsListCmd(apiClientFunc func() (*client.ServiceAPI, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serviceSecretsList",
		Short:   `Retrieves service secrets`,
		Aliases: []string{"secrets"},
		RunE:    runServiceSecretsList(apiClientFunc),
	}

	_ = cmd.PersistentFlags().String(serviceIdFlag, "", "Required. Service id")
	_ = cmd.MarkFlagRequired(serviceIdFlag)

	return cmd
}

// runServiceAdd uses cmd flags to call endpoint api
func runServiceSecretsList(apiClientFunc func() (*client.ServiceAPI, error)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		apiClient, err := apiClientFunc()
		if err != nil {
			return err
		}

		// retrieve flag values from cmd and fill params
		params := service.NewServiceSecretsListParams()

		if value, err := getString(cmd, serviceIdFlag); err != nil {
			return err
		} else if value != nil {
			params.ServiceID = *value
		}

		var formatResponse format
		if value, err := getString(cmd, formatFlag); err != nil {
			return err
		} else if value != nil {
			formatResponse = format(*value)
		}

		if dryRun {
			logDebugf("Params: %+v", params)
			logDebugf("dry-run flag specified. Skip sending request.")
			return nil
		}

		// make request and then print result
		response, err := apiClient.Service.ServiceSecretsList(params,
			openapiClient.APIKeyAuth("X-Token", "header", viper.GetString(tokenFlag)))
		if err != nil {
			return humanizeError(err)
		}

		payload := response.GetPayload()
		if isDefaultPrintFormat(formatResponse) {
			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.SetHeader([]string{"ID", "Value"})
			table.SetBorder(false)
			for _, item := range payload {
				table.Append([]string{item.ID, item.Value})
			}
			table.Render()
		} else {
			return printResult(cmd, formatResponse, payload)
		}
		return nil
	}
}
