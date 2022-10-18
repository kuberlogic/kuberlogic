package app

import (
	"context"
	"github.com/go-errors/errors"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestServiceArchive(t *testing.T) {
	serviceID := "one"

	archived := &v1alpha1.KuberLogicService{
		ObjectMeta: v1.ObjectMeta{
			Name: serviceID,
		},
		Spec: v1alpha1.KuberLogicServiceSpec{
			Type:     "docker-compose",
			Replicas: 1,
			Archived: true,
		},
	}
	archived.MarkArchived()

	cases := []testCase{
		{
			name:   "ok",
			status: 200,
			objects: []runtime.Object{
				&v1alpha1.KuberLogicService{
					ObjectMeta: v1.ObjectMeta{
						Name: serviceID,
					},
					Spec: v1alpha1.KuberLogicServiceSpec{
						Type:     "docker-compose",
						Replicas: 1,
					},
				},
				&v1alpha1.KuberlogicServiceBackup{
					ObjectMeta: v1.ObjectMeta{
						Name: "should-removed",
					},
					Spec: v1alpha1.KuberlogicServiceBackupSpec{
						KuberlogicServiceName: serviceID,
					},
				},
				&v1alpha1.KuberlogicServiceBackup{
					ObjectMeta: v1.ObjectMeta{
						Name: "backup-from-another-service",
					},
					Spec: v1alpha1.KuberlogicServiceBackupSpec{
						KuberlogicServiceName: "another-service",
					},
				},
			},
			result: nil,
			params: apiService.ServiceArchiveParams{
				HTTPRequest: &http.Request{},
				ServiceID:   serviceID,
			},
			helpers: []func(args ...interface{}) error{
				markBackupAsSuccessful,
				checkPreviousBackupsWasRemoved,
				checkServiceIsArchived,
			},
		},
		{
			name:   "already-archived",
			status: 503,
			objects: []runtime.Object{
				archived,
			},
			result: &models.Error{
				Message: "service already is in archive state: one",
			},
			params: apiService.ServiceArchiveParams{
				HTTPRequest: &http.Request{},
				ServiceID:   serviceID,
			},
		},
		{
			name:    "service-not-found",
			status:  404,
			objects: []runtime.Object{},
			result: &models.Error{
				Message: "kuberlogic service not found: one",
			},
			params: apiService.ServiceArchiveParams{
				HTTPRequest: &http.Request{},
				ServiceID:   serviceID,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newFakeHandlers(t, tc.objects...)
			wg := sync.WaitGroup{}
			wg.Add(2)
			go func() {
				defer wg.Done()
				for _, c := range tc.helpers {
					err := c(t, h, serviceID)
					if err != nil {
						t.Error(err)
						return
					}
				}
			}()
			go func() {
				defer wg.Done()
				checkResponse(h.ServiceArchiveHandler(tc.params.(apiService.ServiceArchiveParams), nil), t, tc.status, tc.result)
			}()
			wg.Wait()
		})
	}
}

func markBackupAsSuccessful(args ...interface{}) error {
	t := args[0].(*testing.T)
	h := args[1].(*FakeHandlers)
	serviceID := args[2].(string)
	time.Sleep(time.Second) // waiting backup to be created

	r, err := h.Backups().List(context.TODO(), h.ListOptionsByKeyValue(util.BackupRestoreServiceField, &serviceID))
	if err != nil {
		return err
	}
	if len(r.Items) == 0 {
		return errors.New("successful backup is not found")
	}
	for _, b := range r.Items {
		t.Logf("mark backup %s as successful", b.GetName())
		b.MarkSuccessful()
		gvk := schema.GroupVersionResource{Group: "kuberlogic.com", Version: "v1alpha1", Resource: "kuberlogicservicebackups"}
		err = h.Tracker().Update(gvk, &b, b.GetNamespace())
		if err != nil {
			return err
		}
	}
	return nil
}

func checkServiceIsArchived(args ...interface{}) error {
	h := args[1].(*FakeHandlers)
	serviceID := args[2].(string)
	time.Sleep(time.Second) // waiting service is archived

	s, err := h.Services().Get(context.TODO(), serviceID, v1.GetOptions{})
	if err != nil {
		return err
	}
	if s.Spec.Archived != true {
		return errors.New("service is not archived")
	}
	return nil
}

func checkPreviousBackupsWasRemoved(args ...interface{}) error {
	t := args[0].(*testing.T)
	h := args[1].(*FakeHandlers)
	serviceID := args[2].(string)
	time.Sleep(time.Second) // waiting backups are deleted

	r, err := h.Backups().List(context.TODO(), h.ListOptionsByKeyValue(util.BackupRestoreServiceField, &serviceID))
	if err != nil {
		return err
	}
	for _, b := range r.Items {
		if !strings.HasPrefix(b.GetName(), serviceID) {
			return errors.Errorf("backup is not deleted: %s", b.GetName())
		}
	}

	// check the backup with another service is not deleted
	another, err := h.Backups().Get(context.TODO(), "backup-from-another-service", v1.GetOptions{})
	if err != nil {
		return err
	}
	t.Logf(`backup "%s" is not deleted`, another.GetName())

	return nil
}
