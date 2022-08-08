package cli

import (
	"fmt"
	client2 "github.com/go-openapi/runtime/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"

	"github.com/spf13/cobra"
)

// makeServiceEditCmd returns a cmd to handle operation serviceEdit
func makeServiceEditCmd(apiClientFunc func() (*client.ServiceAPI, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serviceEdit",
		Short:   `Edit service object`,
		Aliases: []string{"edit"},
		RunE:    runServiceEdit(apiClientFunc),
	}

	_ = cmd.PersistentFlags().String(idFlag, "", "service id [required]")
	_ = cmd.PersistentFlags().Int64("replicas", 1, "how many replicas need for service")
	_ = cmd.PersistentFlags().String("version", "", "what the version of service")
	_ = cmd.PersistentFlags().String("domain", "", "domain for external connection to service")
	//_ = cmd.PersistentFlags().String("volume_size", "", "")
	_ = cmd.PersistentFlags().String("backup_schedule", "", "backup schedule in cron format")
	_ = cmd.PersistentFlags().Bool("tls_enabled", false, "")

	// limits
	_ = cmd.PersistentFlags().String("limits.cpu", "", "cpu limits")
	_ = cmd.PersistentFlags().String("limits.memory", "", "memory limits")
	_ = cmd.PersistentFlags().String("limits.volume_size", "", "volume size limits")

	// advanced conf

	return cmd
}

// runServiceAEdit uses cmd flags to call endpoint api
func runServiceEdit(apiClientFunc func() (*client.ServiceAPI, error)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		apiClient, err := apiClientFunc()
		if err != nil {
			return err
		}

		// retrieve flag values from cmd and fill params
		editParams := service.NewServiceEditParams()
		getParams := service.NewServiceGetParams()
		svc := models.Service{}
		svc.Limits = new(models.Limits)

		if value, err := getString(cmd, idFlag); err != nil {
			return err
		} else if value != nil {
			svc.ID = value
			getParams.ServiceID = *value
			editParams.ServiceID = *value
		} else {
			return errors.New("ID is not specified")
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

		if value, err := getString(cmd, "domain"); err != nil {
			return err
		} else if value != nil {
			svc.Domain = *value
		}

		//if value, err := getString(cmd, "volume_size"); err != nil {
		//	return err
		//} else if value != nil {
		//	svc.VolumeSize = *value
		//}

		if value, err := getString(cmd, "backup_schedule"); err != nil {
			return err
		} else if value != nil {
			svc.BackupSchedule = *value
		}

		if value, err := getBool(cmd, "tls_enabled"); err != nil {
			return err
		} else if value != nil {
			svc.TLSEnabled = *value
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

		if value, err := getString(cmd, "limits.volume_size"); err != nil {
			return err
		} else if value != nil {
			svc.Limits.VolumeSize = *value
		}

		var formatResponse format
		if value, err := getString(cmd, formatFlag); err != nil {
			return err
		} else if value != nil {
			formatResponse = format(*value)
		}

		editParams.ServiceItem = &svc

		if dryRun {
			logDebugf("edit params: %+v", editParams.ServiceItem)
			logDebugf("dry-run flag specified. Skip sending request.")
			return nil
		}

		// make request and then print result
		getResponse, err := apiClient.Service.ServiceGet(getParams, client2.APIKeyAuth(
			"X-Token", "header", viper.GetString(tokenFlag)))
		if err != nil {
			return humanizeError(err)
		}

		// set required fields
		svc.Type = getResponse.GetPayload().Type

		// make request and then print result
		response, err := apiClient.Service.ServiceEdit(editParams, client2.APIKeyAuth(
			"X-Token", "header", viper.GetString(tokenFlag)))
		if err != nil {
			return humanizeError(err)
		}
		if isDefaultPrintFormat(formatResponse) {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "Service '%s' successfully edited\n", *svc.ID)
			return err
		} else {
			return printResult(cmd, formatResponse, response.GetPayload())
		}
	}
}
