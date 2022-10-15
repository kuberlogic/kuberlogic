package app

import (
	"context"
	"fmt"
	"sort"
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

func (h *handlers) ServiceUnarchiveHandler(params apiService.ServiceUnarchiveParams, _ *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()
	// Check if service exists first
	service, err := h.Services().Get(ctx, params.ServiceID, v1.GetOptions{})
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
	if !service.Archived() {
		msg := fmt.Sprintf("service is not in archive state: %s", service.GetName())
		h.log.Errorw(msg)
		return apiService.NewServiceUnarchiveServiceUnavailable().WithPayload(&models.Error{
			Message: msg,
		})
	}

	// Unarchive the service in background
	go func() {
		if err := h.UnarchiveKuberlogicService(service.GetName()); err != nil {
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
func (h *handlers) UnarchiveKuberlogicService(serviceName string) error {
	ctx := context.Background()

	h.log.Infow("searching successful backup", "serviceName", serviceName)
	opts := h.ListOptionsByKeyValue(util.BackupRestoreServiceField, &serviceName)
	sortBy := func(backups []*v1alpha1.KuberlogicServiceBackup) sort.Interface {
		return sort.Reverse(BackupsByCreation(backups)) // last backup first
	}
	backup, err := h.Backups().FirstSuccessful(ctx, opts, sortBy)
	if err != nil {
		return errors.Wrap(err, "error finding successful backup")
	}

	h.log.Infow("restore from backup", "serviceName", serviceName, "backupId", backup.GetName())
	restore, err := h.Restores().CreateByBackupName(ctx, backup.Name)
	if k8serrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "restore already exists")
	} else if err != nil {
		return errors.Wrap(err, "error creating restore")
	}

	h.log.Infow("waiting for restore is successful", "serviceName", serviceName, "restoreId", restore.GetName())
	timeout := time.Hour * 2 // wait restoring for 2 hours
	_, err = h.Restores().Wait(ctx, restore, restoringIsSuccessful(h.log, restore.GetName()), timeout)
	if err != nil {
		return errors.Wrap(err, "error waiting for service backup")
	}

	h.log.Infow("unarchive service", "serviceName", serviceName)
	_, err = h.Services().Patch(ctx, serviceName, types.MergePatchType, []byte(`{"spec":{"archived":false}}`), v1.PatchOptions{})
	if err != nil {
		return errors.Wrap(err, "error set archive to service")
	}
	return nil
}

type BackupsByCreation []*v1alpha1.KuberlogicServiceBackup

func (b BackupsByCreation) Len() int      { return len(b) }
func (b BackupsByCreation) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b BackupsByCreation) Less(i, j int) bool {
	return b[i].CreationTimestamp.Before(&b[j].CreationTimestamp)
}

func restoringIsSuccessful(log logging.Logger, name string) func(event watch.Event) (bool, error) {
	return func(event watch.Event) (bool, error) {
		if event.Type == watch.Added || event.Type == watch.Modified {
			obj, ok := event.Object.(*v1alpha1.KuberlogicServiceRestore)
			if !ok {
				log.Infow("event is not a KuberlogicServiceRestore", "event", event)
				return false, errors.New("unexpected object type")
			}
			if obj.GetName() != name {
				log.Infow("event is not for the backup we are waiting for -> skipping.", "observed", obj.GetName(), "waiting", name)
				return false, nil
			}
			if !obj.IsSuccessful() {
				log.Infow("restoring still is not successful", "status", obj.Status.Phase)
				return false, nil
			}
			log.Infow("restoring finally is successful", "status", obj.Status.Phase)
			return true, nil
		}
		log.Debugw("unknown event", "type", event.Type)
		return false, nil
	}
}
