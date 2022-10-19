package cli

import (
	"fmt"
	"strings"

	client2 "github.com/go-openapi/runtime/client"
	"github.com/spf13/viper"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client/logs"

	"github.com/spf13/cobra"
)

const (
	containerNameFlag = "container"
)

// makeLogsCmd returns a cmd to handle operation logArchive
func makeLogsCmd(apiClientFunc func() (*client.ServiceAPI, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: `Show kuberlogic logs`,
		RunE:  runLogsList(apiClientFunc),
	}
	_ = cmd.PersistentFlags().String(containerNameFlag, "", "List logs only for specified container")
	return cmd
}

// runLogsList uses cmd flags to call endpoint api
func runLogsList(apiClientFunc func() (*client.ServiceAPI, error)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		apiClient, err := apiClientFunc()
		if err != nil {
			return err
		}

		params := logs.NewLogListParams()
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
		response, err := apiClient.Logs.LogList(params,
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
