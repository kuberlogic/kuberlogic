package cli

import (
	"fmt"
	"strings"

	client2 "github.com/go-openapi/runtime/client"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client/service"

	"github.com/spf13/cobra"
)

const (
	containerNameFlag = "container"
)

// makeServiceLogsCmd returns a cmd to handle operation serviceLogs
func makeServiceLogsCmd(apiClientFunc func() (*client.ServiceAPI, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serviceLogs",
		Short:   `Show service pod logs`,
		Aliases: []string{"logs"},
		RunE:    runServiceLogsList(apiClientFunc),
	}
	_ = cmd.PersistentFlags().String(serviceIdFlag, "", "Required. Service id")
	_ = cmd.MarkFlagRequired(serviceIdFlag)
	_ = cmd.PersistentFlags().String(containerNameFlag, "", "List logs only for specified container")
	return cmd
}

// runServiceLogsList uses cmd flags to call endpoint api
func runServiceLogsList(apiClientFunc func() (*client.ServiceAPI, error)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		apiClient, err := apiClientFunc()
		if err != nil {
			return err
		}

		params := service.NewServiceLogsListParams()

		if value, err := getString(cmd, serviceIdFlag); err != nil {
			return err
		} else if value != nil {
			params.ServiceID = *value
		} else {
			return errors.New("Service id is required")
		}

		if value, err := getString(cmd, containerNameFlag); err != nil {
			return err
		} else if value != nil {
			params.ContainerName = value
		}

		if dryRun {
			logDebugf("Params: %+v", params.ContainerName)
			logDebugf("dry-run flag specified. Skip sending request.")
			return nil
		}

		// make request and then print result
		response, err := apiClient.Service.ServiceLogsList(params,
			client2.APIKeyAuth("X-Token", "header", viper.GetString(tokenFlag)))
		if err != nil {
			return humanizeError(err)
		}
		output := ""
		for _, container := range response.GetPayload() {
			for _, line := range strings.Split(container.Logs, "\n") {
				output += fmt.Sprintf("%s:\t%s\n", container.ContainerName, line)
			}
		}
		_, err = fmt.Fprint(cmd.OutOrStdout(), output)
		return err
	}
}
