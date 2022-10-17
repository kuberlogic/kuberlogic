package cli

import (
	"fmt"

	client2 "github.com/go-openapi/runtime/client"
	"github.com/spf13/viper"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client/service"

	"github.com/spf13/cobra"
)

// makeServiceArchiveCmd returns a cmd to handle operation serviceArchive
func makeServiceArchiveCmd(apiClientFunc func() (*client.ServiceAPI, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serviceArchive",
		Short:   `Archive service object`,
		Aliases: []string{"archive"},
		RunE:    runServiceArchive(apiClientFunc),
	}

	_ = cmd.PersistentFlags().String(idFlag, "", "Required. Service id")
	_ = cmd.MarkFlagRequired(idFlag)

	return cmd
}

// runServiceArchive uses cmd flags to call endpoint api
func runServiceArchive(apiClientFunc func() (*client.ServiceAPI, error)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		apiClient, err := apiClientFunc()
		if err != nil {
			return err
		}

		// retrieve flag values from cmd and fill params
		params := service.NewServiceArchiveParams()
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
		_, err = apiClient.Service.ServiceArchive(params,
			client2.APIKeyAuth("X-Token", "header", viper.GetString(tokenFlag)))
		if err != nil {
			return humanizeError(err)
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Request for archive service '%s' has been sent\n", params.ServiceID)
		return err
	}
}
