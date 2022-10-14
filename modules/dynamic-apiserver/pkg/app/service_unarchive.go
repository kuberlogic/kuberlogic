package app

import (
	"context"
	"fmt"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func (h *handlers) ServiceUnarchiveHandler(params apiService.ServiceUnarchiveParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()
	// Check if service exists first
	kls, err := h.Services().Get(ctx, params.ServiceID, v1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		msg := fmt.Sprintf("kuberlogic service not found: %s", params.ServiceID)
		h.log.Errorw(msg, "error", err)
		return apiService.NewServiceUnarchiveNotFound().WithPayload(&models.Error{
			Message: msg,
		})
	} else if err != nil {
		msg := "error finding service"
		h.log.Errorw(msg, "error", err)
		return apiService.NewServiceUnarchiveServiceUnavailable().WithPayload(&models.Error{
			Message: msg,
		})
	}
	if !kls.Archived() {
		msg := fmt.Sprintf("service is not in archive state: %s", kls.GetName())
		h.log.Errorw(msg)
		return apiService.NewServiceUnarchiveServiceUnavailable().WithPayload(&models.Error{
			Message: msg,
		})
	}

	// Unarchive the service in background
	go func() {
		if err := h.UnarchiveKuberlogicService(kls); err != nil {
			h.log.Errorw("error unarchiving service", "error", err)
		}
	}()
	return apiService.NewServiceUnarchiveOK()
}

/*
	Unarchive service will:
	1. Find the latest backup
	2. Restoring from the backup
	3. Waiting when the restore completed
	4. Unset "Archive" for the service
*/
func (h *handlers) UnarchiveKuberlogicService(service *v1alpha1.KuberLogicService) error {
	ctx := context.Background()
	serviceId := &service.Name

	h.log.Infow("find successful backup", "serviceId", *serviceId)
	backup, err := h.Backups().GetEarliestSuccessful(ctx, serviceId)
	if err != nil {
		return errors.Wrap(err, "error finding successful backup")
	}

	klr, err := util.RestoreToKuberlogic(&models.Restore{
		ID:       fmt.Sprintf("%s-%d", backup.GetName(), time.Now().Unix()),
		BackupID: backup.GetName(),
	}, backup)
	if err != nil {
		return errors.Wrap(err, "error creating restore object")
	}

	h.log.Infow("restore from backup", "serviceId", *serviceId, "backupId", backup.GetName())
	if _, err := h.Restores().Create(ctx, klr, v1.CreateOptions{}); k8serrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "restore already exists")
	} else if err != nil {
		return errors.Wrap(err, "error creating restore")
	}

	maxRetries := 13 // equals 1 2 4 8 16 32 64 128 256 512 1024 2048 4096
	h.log.Infow("Waiting for restore is successful", "serviceId", *serviceId, "restoreId", klr.GetName())
	err = h.Restores().Wait(ctx, h.log, klr.GetName(), maxRetries)
	if err != nil {
		return errors.Wrap(err, "error waiting for service backup")
	}

	h.log.Infow("unarchive service", "serviceId", *serviceId)
	_, err = h.Services().Patch(ctx, *serviceId, types.MergePatchType, []byte(`{"spec":{"archive":false}}`), v1.PatchOptions{})
	if err != nil {
		return errors.Wrap(err, "error set archive to service")
	}
	return nil
}
