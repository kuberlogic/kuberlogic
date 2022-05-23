package cli

import (
	"github.com/go-openapi/runtime"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/client/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/models"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"strconv"

	"github.com/spf13/cobra"
)

// makeServiceListCmd returns a cmd to handle operation serviceList
func makeServiceListCmd(apiClientFunc func() (*client.ServiceAPI, error)) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "serviceList",
		Short:   `List of service objects`,
		Aliases: []string{"list"},
		RunE:    runServiceList(apiClientFunc),
	}

	return cmd, nil
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

		payload, err := parseServiceListResult(apiClient.Service.ServiceList(params))
		if err != nil {
			return err
		}

		if isDefaultPrintFormat(formatResponse) {
			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.SetHeader([]string{"№", "ID", "Type", "Replica", "Version", "Host", "Status"})
			table.SetBorder(false)
			for i, item := range payload {
				table.Append([]string{
					strconv.Itoa(i), *item.ID, *item.Type, strconv.Itoa(int(*item.Replicas)),
					item.Version, item.Host, item.Status,
				})
			}
			table.Render()
		} else {
			return printResult(cmd, formatResponse, payload)
		}
		return nil
	}
}

// parseServiceListResult parses request result and return the string content
func parseServiceListResult(resp *service.ServiceListOK, respErr error) (models.Services, error) {
	if respErr != nil {
		switch respErr.(type) {
		case *service.ServiceListBadRequest:
			err := respErr.(*service.ServiceListBadRequest)
			return nil, errors.Errorf(err.Payload.Message)
		case *service.ServiceListServiceUnavailable:
			return nil, errors.Errorf("Service unavailable [%v]", respErr)
		case *service.ServiceListUnauthorized:
			return nil, errors.Errorf("Unauthorized [%v]", respErr)
		case *service.ServiceListForbidden:
			return nil, errors.Errorf("Forbidden [%v]", respErr)
		case *service.ServiceListUnprocessableEntity:
			err := respErr.(*service.ServiceListUnprocessableEntity)
			return nil, errors.Errorf("validation error: %s", err.Payload.Message)
		case *runtime.APIError:
			return nil, errors.Errorf("APIError [%v]", respErr)
		default:
			return nil, errors.Errorf("Unknown response type: %T [%v]", respErr, respErr)
		}

	}
	return resp.Payload, nil
}
