package app

import (
	"context"
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func (h *handlers) ServiceArchiveHandler(params apiService.ServiceArchiveParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	// Check if service exists first
	kls, err := h.Services().Get(ctx, params.ServiceID, v1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		msg := fmt.Sprintf("kuberlogic service not found: %s", params.ServiceID)
		h.log.Errorw(msg, "error", err)
		return apiService.NewServiceArchiveNotFound().WithPayload(&models.Error{
			Message: msg,
		})
	} else if err != nil {
		msg := "error finding service"
		h.log.Errorw(msg, "error", err)
		return apiService.NewServiceArchiveServiceUnavailable().WithPayload(&models.Error{
			Message: msg,
		})
	}
	if kls.Archived() {
		msg := fmt.Sprintf("service already is in archive state: %s", kls.GetName())
		h.log.Errorw(msg)
		return apiService.NewServiceArchiveServiceUnavailable().WithPayload(&models.Error{
			Message: msg,
		})
	}

	// Get service archive in background
	go func() {
		if err := h.ArchiveKuberlogicService(kls); err != nil {
			h.log.Errorw("error archiving service", "error", err)
		}
	}()
	return apiService.NewServiceArchiveOK()
}

/*
	Archive service will:
	1. Take new backup of a service (if backups enabled)
	2. Waiting the backup is done
	3. Remove all previous backups
	4. Set "Archive" for the service
*/
func (h *handlers) ArchiveKuberlogicService(service *v1alpha1.KuberLogicService) error {
	ctx := context.Background()
	serviceId := &service.Name

	h.log.Infow("Taking backup of the service", "serviceId", *serviceId)
	archive, err := h.Backups().CreateBackupByServiceName(ctx, *serviceId)
	if err != nil {
		return errors.Wrap(err, "error creating service backup")
	}

	archiveID := archive.GetName()
	maxRetries := 13 // equals 1 2 4 8 16 32 64 128 256 512 1024 2048 4096
	h.log.Infow("Waiting for backup to be ready", "serviceId", *serviceId, "backupId", archiveID)
	err = h.Backups().Wait(ctx, h.log, serviceId, archiveID, maxRetries)
	if err != nil {
		return errors.Wrap(err, "error waiting for service backup")
	}

	h.log.Infow("Deleting previous backups", "serviceId", *serviceId)
	r, err := h.Backups().ListByServiceName(ctx, serviceId)
	if err != nil {
		return errors.Wrap(err, "error listing service backups")
	}
	for _, backup := range r.Items {
		if backup.GetName() == archiveID {
			continue
		}
		h.log.Infow("Deleting backup", "serviceId", *serviceId, "backupId", backup.GetName())
		if err = h.Backups().Delete(ctx, backup.Name, v1.DeleteOptions{}); err != nil {
			return errors.Wrap(err, "error deleting service backup")
		}
	}

	h.log.Infow("Archive service", "serviceId", *serviceId)
	_, err = h.Services().Patch(ctx, *serviceId, types.MergePatchType, []byte(`{"spec":{"archive":true}}`), v1.PatchOptions{})
	if err != nil {
		return errors.Wrap(err, "error set archive to service")
	}
	return nil
}
