package cli

import (
	"fmt"

	client2 "github.com/go-openapi/runtime/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"
	"github.com/spf13/viper"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client/service"

	"github.com/spf13/cobra"
)

func makeServiceUnarchiveCmd(apiClientFunc func() (*client.ServiceAPI, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serviceUnarchive",
		Short:   `Unarchive service object`,
		Aliases: []string{"unarchive"},
		RunE:    runServiceUnarchive(apiClientFunc),
	}

	_ = cmd.PersistentFlags().String(idFlag, "", "Required. Service ID.")
	_ = cmd.MarkFlagRequired(idFlag)
	return cmd
}

func runServiceUnarchive(apiClientFunc func() (*client.ServiceAPI, error)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		apiClient, err := apiClientFunc()
		if err != nil {
			return err
		}

		// retrieve flag values from cmd and fill params
		params := service.NewServiceUnarchiveParams()
		if value, err := getString(cmd, idFlag); err != nil {
			return err
		} else if value != nil {
			params.ServiceID = *value
		}

		if dryRun {
			logDebugf("Params: %+v", params.ServiceID)
			logDebugf("dry-run flag specified. Skip sending request.")
			return nil
		}

		// make request and then print result
		_, err = apiClient.Service.ServiceUnarchive(params,
			client2.APIKeyAuth("X-Token", "header", viper.GetString(tokenFlag)))
		if err != nil {
			return humanizeError(err)
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Request for archive service '%s' has been sent\n", params.ServiceID)
		return err
	}
}
