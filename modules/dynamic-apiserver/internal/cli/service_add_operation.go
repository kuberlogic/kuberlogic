package cli

import (
	"encoding/json"
	"fmt"
	"github.com/go-openapi/runtime"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/client/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/models"

	"github.com/spf13/cobra"
)

// makeServiceAddCmd returns a cmd to handle operation serviceAdd
func makeServiceAddCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "serviceAdd",
		Short:   `Adds service object`,
		Aliases: []string{"add"},
		RunE:    runServiceAdd,
	}

	_ = cmd.PersistentFlags().String("name", "", "name of service")
	_ = cmd.PersistentFlags().String("namespace", "", "namespace for service")
	_ = cmd.PersistentFlags().String("type", "", "type of service")
	_ = cmd.PersistentFlags().Int64("replicas", 0, "how many replicas need for service")
	_ = cmd.PersistentFlags().String("version", "", "what the version of service")
	_ = cmd.PersistentFlags().String("volumeSize", "", "")

	// limits
	_ = cmd.PersistentFlags().String("limits.cpu", "", "cpu limits")
	_ = cmd.PersistentFlags().String("limits.memory", "", "memory limits")
	_ = cmd.PersistentFlags().String("limits.volumeSize", "", "volume size limits")

	// advanced conf

	return cmd, nil
}

// runServiceAdd uses cmd flags to call endpoint api
func runServiceAdd(cmd *cobra.Command, args []string) error {
	var err error
	appCli, err := makeClient(cmd, args)
	if err != nil {
		return err
	}
	// retrieve flag values from cmd and fill params
	params := service.NewServiceAddParams()
	svc := models.Service{}
	svc.Limits = new(models.Limits)

	if value, err := getString(cmd, "name"); err != nil {
		return err
	} else if value != nil {
		svc.Name = value
	}

	if value, err := getString(cmd, "namespace"); err != nil {
		return err
	} else if value != nil {
		svc.Ns = *value
	}

	if value, err := getString(cmd, "type"); err != nil {
		return err
	} else if value != nil {
		svc.Type = value
	}

	if value, err := setInt64(cmd, "replicas"); err != nil {
		return err
	} else if value != nil {
		svc.Replicas = value
	}

	if value, err := getString(cmd, "version"); err != nil {
		return err
	} else if value != nil {
		svc.Version = *value
	}

	if value, err := getString(cmd, "volumeSize"); err != nil {
		return err
	} else if value != nil {
		svc.VolumeSize = *value
	}

	if value, err := getString(cmd, "limits.cpu"); err != nil {
		return err
	} else if value != nil {
		svc.Limits.CPU = *value
	}

	if value, err := getString(cmd, "limits.memory"); err != nil {
		return err
	} else if value != nil {
		svc.Limits.Memory = *value
	}

	if value, err := getString(cmd, "limits.volumeSize"); err != nil {
		return err
	} else if value != nil {
		svc.Limits.VolumeSize = *value
	}

	params.ServiceItem = &svc
	if dryRun {
		logDebugf("Params: %+v", params.ServiceItem)
		logDebugf("dry-run flag specified. Skip sending request.")
		return nil
	}
	// make request and then print result
	payload, err := parseServiceAddResult(appCli.Service.ServiceAdd(params))
	if err != nil {
		return err
	}
	return printResult(payload)
}

func printResult(payload *models.Service) error {
	var result interface{}
	var err error
	switch formatResponse {
	case jsonFormat:
		result, err = json.MarshalIndent(payload, "", "\t")
	case yamlFormat:
		result, err = yaml.Marshal(payload)
	default:
		result = fmt.Sprintf("Service '%s' successfully created\n", *payload.Name)
	}

	if err != nil {
		return err
	}
	fmt.Printf("%s\n", result)
	return nil
}

// parseServiceAddResult parses request result and return the string content
func parseServiceAddResult(resp *service.ServiceAddCreated, respErr error) (*models.Service, error) {
	if respErr != nil {
		switch respErr.(type) {
		case *service.ServiceAddBadRequest:
			err := respErr.(*service.ServiceAddBadRequest)
			return nil, errors.Errorf(err.Payload.Message)
		case *service.ServiceAddServiceUnavailable:
			return nil, errors.Errorf("Service unavailable [%v]", respErr)
		case *service.ServiceAddUnauthorized:
			return nil, errors.Errorf("Unauthorized [%v]", respErr)
		case *service.ServiceAddForbidden:
			return nil, errors.Errorf("Forbidden [%v]", respErr)
		case *service.ServiceAddConflict:
			return nil, errors.Errorf("Conflict [%v]", respErr)
		case *service.ServiceAddUnprocessableEntity:
			err := respErr.(*service.ServiceAddUnprocessableEntity)
			return nil, errors.Errorf("validation error: %s", err.Payload.Message)
		case *runtime.APIError:
			return nil, errors.Errorf("APIError [%v]", respErr)
		default:
			return nil, errors.Errorf("Unknown response type: %T [%v]", respErr, respErr)
		}

	}
	return resp.Payload, nil
}
