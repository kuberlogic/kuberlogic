package cli

import (
	"fmt"

	client2 "github.com/go-openapi/runtime/client"
	"github.com/spf13/viper"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"

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

	_ = cmd.PersistentFlags().String(idFlag, "", "Required. Service id")
	_ = cmd.MarkFlagRequired(idFlag)
	_ = cmd.PersistentFlags().String("type", "", "Required. Supported service type")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.PersistentFlags().Int64("replicas", 1, "Service replicas count")
	_ = cmd.PersistentFlags().String("version", "", "Service version")
	_ = cmd.PersistentFlags().String("backup_schedule", "", "Backup schedule in cron format")
	_ = cmd.PersistentFlags().String("domain", "", "Custom domain for a service")
	_ = cmd.PersistentFlags().Bool("insecure", false, "Use HTTP protocol instead of HTTPS")
	_ = cmd.PersistentFlags().String(subscriptionId, "", "Subscription ID")
 	_ = cmd.PersistentFlags().Bool("use_letsencrypt", false, "use Let's Encrypt for service as TLS certificate issuer")

	// limits
	_ = cmd.PersistentFlags().String("limits.cpu", "", "CPU limits")
	_ = cmd.PersistentFlags().String("limits.memory", "", "Memory limits")
	_ = cmd.PersistentFlags().String("limits.storage", "", "Storage limits")

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

		if value, err := getBool(cmd, "use_letsencrypt"); err != nil {
			return err
		} else if value != nil {
			svc.UseLetsencrypt = *value
		}

		if value, err := getString(cmd, "domain"); err != nil {
			return err
		} else if value != nil {
			svc.Domain = *value
		}

		if value, err := getString(cmd, subscriptionId); err != nil {
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
