package cli

import (
	client2 "github.com/go-openapi/runtime/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strconv"
)

// makeServiceExplainCmd returns a cmd to handle operation serviceLogs
func makeServiceExplainCmd(apiClientFunc func() (*client.ServiceAPI, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serviceExplain",
		Short:   `Explain status of service`,
		Aliases: []string{"explain"},
		RunE:    runServiceExplain(apiClientFunc),
	}
	_ = cmd.PersistentFlags().String(serviceIdFlag, "", "Required. Service id")
	_ = cmd.MarkFlagRequired(serviceIdFlag)
	return cmd
}

// runServiceExplain uses cmd flags to call endpoint api
func runServiceExplain(apiClientFunc func() (*client.ServiceAPI, error)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		apiClient, err := apiClientFunc()
		if err != nil {
			return err
		}

		params := service.NewServiceExplainParams()

		if value, err := getString(cmd, serviceIdFlag); err != nil {
			return err
		} else if value != nil {
			params.ServiceID = *value
		} else {
			return errors.New("Service id is required")
		}
		var formatResponse format
		if value, err := getString(cmd, formatFlag); err != nil {
			return err
		} else if value != nil {
			formatResponse = format(*value)
		}

		if dryRun {
			logDebugf("Params: %+v", params.ServiceID)
			logDebugf("dry-run flag specified. Skip sending request.")
			return nil
		}

		// make request and then print result
		response, err := apiClient.Service.ServiceExplain(params,
			client2.APIKeyAuth("X-Token", "header", viper.GetString(tokenFlag)))
		if err != nil {
			return humanizeError(err)
		}
		if !isDefaultPrintFormat(formatResponse) {
			return printResult(cmd, formatResponse, response.GetPayload())
		}

		payload := response.GetPayload()

		if payload.Ingress.Error != "" {
			cmd.Printf("Ingress error: %s", payload.Ingress.Error)
		} else {
			printIngress(cmd, payload.Ingress)
		}

		if payload.Pod.Error != "" {
			cmd.Printf("Container errors: %s\n", payload.Pod.Error)
		} else {
			printContainers(cmd, payload.Pod)
		}

		if payload.Pvc.Error != "" {
			cmd.Printf("PVC error: %s\n", payload.Pvc.Error)
		} else {
			printPvc(cmd, payload.Pvc)
		}
		return nil
	}
}

func printContainers(cmd *cobra.Command, data *models.ExplainPod) {
	table := tablewriter.NewWriter(cmd.OutOrStdout())

	table.SetHeader([]string{"№", "NAME", "STATUS", "RESTART COUNT"})

	table.SetBorder(false)
	for i, item := range data.Containers {
		table.Append([]string{
			strconv.Itoa(i + 1), item.Name, item.Status, strconv.FormatInt(*item.RestartCount, 10),
		})
	}
	cmd.Println("Containers:")
	table.Render()
	cmd.Println()
}

func printPvc(cmd *cobra.Command, data *models.ExplainPvc) {
	table := tablewriter.NewWriter(cmd.OutOrStdout())

	table.SetHeader([]string{"PHASE", "SIZE", "STORAGE CLASS"})
	table.SetBorder(false)
	table.Append([]string{
		data.Phase, data.Size, data.StorageClass,
	})

	cmd.Println("Storage:")
	table.Render()
	cmd.Println()
}

func printIngress(cmd *cobra.Command, data *models.ExplainIngress) {
	table := tablewriter.NewWriter(cmd.OutOrStdout())

	table.SetHeader([]string{"№", "HOST", "INGRESS CLASS"})
	table.SetBorder(false)
	for i, item := range data.Hosts {
		table.Append([]string{
			strconv.Itoa(i + 1), item, data.IngressClass,
		})
	}
	cmd.Println("Ingress:")
	table.Render()
	cmd.Println()
}
