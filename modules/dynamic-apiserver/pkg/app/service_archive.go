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
	"k8s.io/apimachinery/pkg/watch"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/logging"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
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
		if err := h.archiveService(kls.GetName()); err != nil {
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
func (h *handlers) archiveService(serviceName string) error {
	ctx := context.Background()

	h.log.Infow("taking backup of the service", "serviceName", serviceName)
	backup, err := h.Backups().CreateByServiceName(ctx, serviceName)
	if err != nil {
		return errors.Wrap(err, "error creating service backup")
	}

	h.log.Infow("waiting for backup to be ready", "serviceName", serviceName, "backupName", backup.GetName())
	timeout := time.Hour * 2 // wait backup for 2 hours
	_, err = h.Backups().Wait(ctx, backup, backupIsSuccessful(h.log, backup.GetName()), timeout)
	if err != nil {
		return errors.Wrap(err, "error waiting for service backup")
	}

	h.log.Infow("deleting previous backups", "serviceName", serviceName)
	opts := h.ListOptionsByKeyValue(util.BackupRestoreServiceField, serviceName)
	r, err := h.Backups().List(ctx, opts)
	if err != nil {
		return errors.Wrap(err, "error listing service backups")
	}
	for _, item := range r.Items {
		if item.GetName() == backup.GetName() {
			continue
		}
		h.log.Infow("deleting backup", "serviceName", serviceName, "backupName", item.GetName())
		if err = h.Backups().Delete(ctx, item.Name, v1.DeleteOptions{}); err != nil {
			return errors.Wrap(err, "error deleting service backup")
		}
	}

	h.log.Infow("archive service", "serviceName", serviceName)
	_, err = h.Services().Patch(ctx, serviceName, types.MergePatchType, []byte(`{"spec":{"archived":true}}`), v1.PatchOptions{})
	if err != nil {
		return errors.Wrap(err, "error set backup to service")
	}
	return nil
}

func backupIsSuccessful(log logging.Logger, name string) func(event watch.Event) (bool, error) {
	return func(event watch.Event) (bool, error) {
		if event.Type == watch.Added || event.Type == watch.Modified {
			obj, ok := event.Object.(*v1alpha1.KuberlogicServiceBackup)
			if !ok {
				log.Infow("event is not a KuberlogicServiceBackup", "event", event)
				return false, errors.New("unexpected object type")
			}
			if obj.GetName() != name {
				log.Infow("event is not for the backup we are waiting for -> skipping.", "observed", obj.GetName(), "waiting", name)
				return false, nil
			}
			if !obj.IsSuccessful() {
				log.Infow("backup still is not successful", "status", obj.Status.Phase)
				return false, nil
			}
			log.Infow("backup finally is successful", "status", obj.Status.Phase)
			return true, nil
		}
		log.Debugw("unknown event", "type", event.Type)
		return false, nil
	}
}
