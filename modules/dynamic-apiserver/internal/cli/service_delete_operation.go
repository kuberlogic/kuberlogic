package cli

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"github.com/go-openapi/runtime"
	client2 "github.com/go-openapi/runtime/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/client/service"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

// makeServiceDeleteCmd returns a cmd to handle operation serviceDelete
func makeServiceDeleteCmd(apiClientFunc func() (*client.ServiceAPI, error)) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "serviceDelete",
		Short:   `Deletes a service object`,
		Aliases: []string{"delete"},
		RunE:    runServiceDelete(apiClientFunc),
	}

	_ = cmd.PersistentFlags().String("id", "", "service id")

	return cmd, nil
}

// runServiceDelete uses cmd flags to call endpoint api
func runServiceDelete(apiClientFunc func() (*client.ServiceAPI, error)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {

		apiClient, err := apiClientFunc()
		if err != nil {
			return err
		}

		// retrieve flag values from cmd and fill params
		params := service.NewServiceDeleteParams()

		var id string
		if value, err := getString(cmd, "id"); err != nil {
			return err
		} else if value != nil {
			id = *value
		}

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

		payload, e := apiClient.Service.ServiceDelete(params, client2.APIKeyAuth("X-Token", "header", viper.GetString("token")))
		// make request and then print result
		if err = parseServiceDeleteResult(id, e); err != nil {
			return err
		}
		if isDefaultPrintFormat(formatResponse) {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "Service '%s' successfully removed\n", id)
			return err
		} else {
			return printResult(cmd, formatResponse, payload)
		}
	}
}

// parseServiceListResult parses request result and return the string content
func parseServiceDeleteResult(id string, respErr error) error {
	if respErr != nil {
		switch respErr.(type) {
		case *service.ServiceDeleteBadRequest:
			err := respErr.(*service.ServiceDeleteBadRequest)
			return errors.Errorf(err.Payload.Message)
		case *service.ServiceDeleteServiceUnavailable:
			return errors.Errorf("Service unavailable [%v]", respErr)
		case *service.ServiceDeleteUnauthorized:
			return errors.Errorf("Unauthorized [%v]", respErr)
		case *service.ServiceDeleteForbidden:
			return errors.Errorf("Forbidden [%v]", respErr)
		case *service.ServiceDeleteNotFound:
			return errors.Errorf("Record not found with id '%s'", id)
		case *service.ServiceDeleteUnprocessableEntity:
			//err := respErr.(*service.ServiceDeleteUnprocessableEntity)
			return errors.Errorf("validation error: %s", "")
		case *runtime.APIError:
			return errors.Errorf("APIError [%v]", respErr)
		default:
			return errors.Errorf("Unknown response type: %T [%v]", respErr, respErr)
		}
	}
	return nil
}
