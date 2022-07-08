package cli

import (
	"fmt"
	client2 "github.com/go-openapi/runtime/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"
	"github.com/spf13/viper"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"

	"github.com/spf13/cobra"
)

// makeServiceAddCmd returns a cmd to handle operation serviceAdd
func makeServiceAddCmd(apiClientFunc func() (*client.ServiceAPI, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serviceAdd",
		Short:   `Adds service object`,
		Aliases: []string{"add"},
		RunE:    runServiceAdd(apiClientFunc),
	}

	_ = cmd.PersistentFlags().String(id_flag, "", "service id")
	_ = cmd.PersistentFlags().String("type", "", "type of service")
	_ = cmd.PersistentFlags().Int64("replicas", 1, "how many replicas need for service")
	_ = cmd.PersistentFlags().String("version", "", "what the version of service")
	_ = cmd.PersistentFlags().String("domain", "", "domain for external connection to service")
	_ = cmd.PersistentFlags().String("volume_size", "", "")
	_ = cmd.PersistentFlags().Bool("tls_enabled", false, "")
	_ = cmd.PersistentFlags().Bool(subscription_id_flag, false, "")

	// limits
	_ = cmd.PersistentFlags().String("limits.cpu", "", "cpu limits")
	_ = cmd.PersistentFlags().String("limits.memory", "", "memory limits")
	_ = cmd.PersistentFlags().String("limits.volume_size", "", "volume size limits")

	// advanced conf

	return cmd
}

// runServiceAdd uses cmd flags to call endpoint api
func runServiceAdd(apiClientFunc func() (*client.ServiceAPI, error)) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error

		apiClient, err := apiClientFunc()
		if err != nil {
			return err
		}

		// retrieve flag values from cmd and fill params
		params := service.NewServiceAddParams()
		svc := models.Service{}
		svc.Limits = new(models.Limits)

		if value, err := getString(cmd, id_flag); err != nil {
			return err
		} else if value != nil {
			svc.ID = value
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

		if value, err := getString(cmd, "domain"); err != nil {
			return err
		} else if value != nil {
			svc.Domain = *value
		}

		if value, err := getString(cmd, "volume_size"); err != nil {
			return err
		} else if value != nil {
			svc.VolumeSize = *value
		}

		if value, err := getBool(cmd, "tls_enabled"); err != nil {
			return err
		} else {
			svc.TLSEnabled = value
		}

		if value, err := getString(cmd, subscription_id_flag); err != nil {
			return err
		} else if value != nil {
			svc.Subscription = *value
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
		if value, err := getString(cmd, "format"); err != nil {
			return err
		} else if value != nil {
			formatResponse = format(*value)
		}

		params.ServiceItem = &svc
		if dryRun {
			logDebugf("Params: %+v", params.ServiceItem)
			logDebugf("dry-run flag specified. Skip sending request.")
			return nil
		}

		// make request and then print result
		response, err := apiClient.Service.ServiceAdd(params,
			client2.APIKeyAuth("X-Token", "header", viper.GetString("token")))
		if err != nil {
			return humanizeError(err)
		}
		payload := response.GetPayload()
		if isDefaultPrintFormat(formatResponse) {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "Service '%s' successfully created\n", *payload.ID)
			return err
		} else {
			return printResult(cmd, formatResponse, payload)
		}

	}
}
