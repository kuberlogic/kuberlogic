package cli

import (
	"fmt"
	"github.com/pkg/errors"

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

func getString(cmd *cobra.Command, flag string) (value *string, err error) {
	if cmd.Flags().Changed(flag) {
		value, err := cmd.Flags().GetString(flag)
		if err != nil {
			return nil, err
		}
		return &value, nil
	}
	return
}

func setInt64(cmd *cobra.Command, flag string) (value *int64, err error) {
	if cmd.Flags().Changed(flag) {
		value, err := cmd.Flags().GetInt64(flag)
		if err != nil {
			return nil, err
		}
		return &value, nil
	}
	return
}

// runServiceAdd uses cmd flags to call endpoint api
func runServiceAdd(cmd *cobra.Command, args []string) error {
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
	fmt.Printf("Service '%s' successfully created\n", *payload.Name)
	return nil
}

// parseServiceAddResult parses request result and return the string content
func parseServiceAddResult(resp *service.ServiceAddCreated, respErr error) (*models.Service, error) {
	if respErr != nil {
		switch respErr.(type) {
		case *service.ServiceAddBadRequest:
			return nil, errors.Errorf("Bad request [%v]", respErr)
		case *service.ServiceAddServiceUnavailable:
			return nil, errors.Errorf("Service unavailable [%v]", respErr)
		case *service.ServiceAddUnauthorized:
			return nil, errors.Errorf("Unauthorized [%v]", respErr)
		case *service.ServiceAddForbidden:
			return nil, errors.Errorf("Forbidden [%v]", respErr)
		case *service.ServiceAddConflict:
			return nil, errors.Errorf("Conflict [%v]", respErr)
		default:
			return nil, errors.Errorf("Unknown response type: %T [%v]", respErr, respErr)
		}

	}
	return resp.Payload, nil
}
