package cli

import (
	"fmt"
	openapiClient "github.com/go-openapi/runtime/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client/service"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"strings"

	"github.com/spf13/cobra"
)

var (
	errCredentialsNotFound       = errors.New("credentials pairs not found")
	errPartialCredentials        = errors.New("failed to parse provided credentials pairs")
	errDuplicateCredentialsPairs = errors.New("duplicate credentials key")
)

// makeServiceAddCmd returns a cmd to handle operation serviceAdd
func makeServiceCredentialsUpdateCmd(apiClientFunc func() (*client.ServiceAPI, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serviceCredentialsUpdate",
		Short:   `Updates a service credentials. Pass credentials in key=value args.`,
		Aliases: []string{"credentials-update"},
		RunE:    runServiceCredentialsUpdate(apiClientFunc),
	}

	_ = cmd.PersistentFlags().String(serviceIdFlag, "", "service id")

	return cmd
}

// runServiceAdd uses cmd flags to call endpoint api
func runServiceCredentialsUpdate(apiClientFunc func() (*client.ServiceAPI, error)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		apiClient, err := apiClientFunc()
		if err != nil {
			return err
		}

		// retrieve flag values from cmd and fill params
		params := service.NewServiceCredentialsUpdateParams()

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

		params.ServiceCredentials = make(map[string]string, len(args))
		if len(args) == 0 {
			return errCredentialsNotFound
		}
		for _, arg := range args {
			credentialsParam := strings.Split(arg, "=")
			if len(credentialsParam) != 2 {
				return errPartialCredentials
			}
			if _, found := params.ServiceCredentials[credentialsParam[0]]; found {
				return errDuplicateCredentialsPairs
			}
			params.ServiceCredentials[credentialsParam[0]] = credentialsParam[1]
		}

		if dryRun {
			logDebugf("Params: %+v", params)
			logDebugf("dry-run flag specified. Skip sending request.")
			return nil
		}

		// make request and then print result
		response, err := apiClient.Service.ServiceCredentialsUpdate(params,
			openapiClient.APIKeyAuth("X-Token", "header", viper.GetString(tokenFlag)))
		if err != nil {
			return humanizeError(err)
		}
		if isDefaultPrintFormat(formatResponse) {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "Credentials updated\n")
			return err
		} else {
			return printResult(cmd, formatResponse, response)
		}
	}
}
