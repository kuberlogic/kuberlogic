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

	_ = cmd.PersistentFlags().String(idFlag, "", "service id")
	_ = cmd.PersistentFlags().String("type", "", "type of service")
	_ = cmd.PersistentFlags().Int64("replicas", 1, "how many replicas need for service")
	_ = cmd.PersistentFlags().String("version", "", "what the version of service")
	_ = cmd.PersistentFlags().String("backup_schedule", "", "backup schedule in cron format")
	_ = cmd.PersistentFlags().String("domain", "", "on which domain service will be available")
	_ = cmd.PersistentFlags().Bool("insecure", false, "setup unsecure service with http, not https")
	_ = cmd.PersistentFlags().Bool(subscriptionIdFlag, false, "")

	// limits
	_ = cmd.PersistentFlags().String("limits.cpu", "", "cpu limits")
	_ = cmd.PersistentFlags().String("limits.memory", "", "memory limits")
	_ = cmd.PersistentFlags().String("limits.storage", "", "storage limits")

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

		if value, err := getString(cmd, idFlag); err != nil {
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

		if value, err := getString(cmd, "backup_schedule"); err != nil {
			return err
		} else if value != nil {
			svc.BackupSchedule = *value
		}

		if value, err := getBool(cmd, "insecure"); err != nil {
			return err
		} else if value != nil {
			svc.Insecure = *value
		}

		if value, err := getString(cmd, "domain"); err != nil {
			return err
		} else if value != nil {
			svc.Domain = *value
		}

		if value, err := getString(cmd, subscriptionIdFlag); err != nil {
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

		if value, err := getString(cmd, "limits.storage"); err != nil {
			return err
		} else if value != nil {
			svc.Limits.Storage = *value
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
			client2.APIKeyAuth("X-Token", "header", viper.GetString(tokenFlag)))
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
