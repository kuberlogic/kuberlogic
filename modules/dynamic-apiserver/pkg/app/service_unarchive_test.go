package app

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

func TestServiceUnarchive(t *testing.T) {
	serviceID := "one"

	unarchived := &v1alpha1.KuberLogicService{
		ObjectMeta: v1.ObjectMeta{
			Name: serviceID,
		},
		Spec: v1alpha1.KuberLogicServiceSpec{
			Type:     "docker-compose",
			Replicas: 1,
		},
	}

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
				archived,
				&v1alpha1.KuberlogicServiceBackup{
					ObjectMeta: v1.ObjectMeta{
						Name: "target",
						Labels: map[string]string{
							util.BackupRestoreServiceField: serviceID,
						},
						CreationTimestamp: v1.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC),
					},
					Spec: v1alpha1.KuberlogicServiceBackupSpec{
						KuberlogicServiceName: serviceID,
					},
					Status: v1alpha1.KuberlogicServiceBackupStatus{
						Phase: "Successful",
					},
				},
				&v1alpha1.KuberlogicServiceBackup{
					ObjectMeta: v1.ObjectMeta{
						Name: "earlier",
						Labels: map[string]string{
							util.BackupRestoreServiceField: serviceID,
						},
						CreationTimestamp: v1.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
					},
					Spec: v1alpha1.KuberlogicServiceBackupSpec{
						KuberlogicServiceName: serviceID,
					},
					Status: v1alpha1.KuberlogicServiceBackupStatus{
						Phase: "Successful",
					},
				},
				&v1alpha1.KuberlogicServiceBackup{
					ObjectMeta: v1.ObjectMeta{
						Name: "not-successful",
						Labels: map[string]string{
							util.BackupRestoreServiceField: serviceID,
						},
						CreationTimestamp: v1.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC),
					},
					Spec: v1alpha1.KuberlogicServiceBackupSpec{
						KuberlogicServiceName: serviceID,
					},
					Status: v1alpha1.KuberlogicServiceBackupStatus{
						Phase: "Failed",
					},
				},
				&v1alpha1.KuberlogicServiceBackup{
					ObjectMeta: v1.ObjectMeta{
						Name: "from-another-service",
						Labels: map[string]string{
							util.BackupRestoreServiceField: "from-another-service",
						},
						CreationTimestamp: v1.Date(2022, 1, 4, 0, 0, 0, 0, time.UTC),
					},
					Spec: v1alpha1.KuberlogicServiceBackupSpec{
						KuberlogicServiceName: "another-service",
					},
					Status: v1alpha1.KuberlogicServiceBackupStatus{
						Phase: "Successful",
					},
				},
			},
			result: nil,
			params: apiService.ServiceUnarchiveParams{
				HTTPRequest: &http.Request{},
				ServiceID:   serviceID,
			},
			helpers: []func(args ...interface{}) error{
				markRestoreAsSuccessful,
				checkServiceIsNotArchived,
			},
		},
		{
			name:   "not-archived",
			status: 503,
			objects: []runtime.Object{
				unarchived,
			},
			result: &models.Error{
				Message: "service is not in archive state: one",
			},
			params: apiService.ServiceUnarchiveParams{
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
			params: apiService.ServiceUnarchiveParams{
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
				checkResponse(h.ServiceUnarchiveHandler(tc.params.(apiService.ServiceUnarchiveParams), nil), t, tc.status, tc.result)
			}()
			wg.Wait()
		})
	}
}

func markRestoreAsSuccessful(args ...interface{}) error {
	t := args[0].(*testing.T)
	h := args[1].(*FakeHandlers)
	time.Sleep(time.Second) // waiting restore object to be created

	r, err := h.Restores().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return err
	}
	if len(r.Items) == 0 {
		return errors.New("successful restore object is not found")
	}
	for _, b := range r.Items {
		backupName := b.Spec.KuberlogicServiceBackup
		if backupName != "target" {
			return errors.Errorf("incorrect backup not found: %s", backupName)
		}
		t.Logf("backup was found: %s", backupName)
		t.Logf("mark restore object %s as successful", b.GetName())
		b.MarkSuccessful()
		gvk := schema.GroupVersionResource{Group: "kuberlogic.com", Version: "v1alpha1", Resource: "kuberlogicservicerestores"}
		err = h.Tracker().Update(gvk, &b, b.GetNamespace())
		if err != nil {
			return err
		}
	}
	return nil
}

func checkServiceIsNotArchived(args ...interface{}) error {
	h := args[1].(*FakeHandlers)
	serviceID := args[2].(string)
	time.Sleep(time.Second) // waiting service is not archived

	s, err := h.Services().Get(context.TODO(), serviceID, v1.GetOptions{})
	if err != nil {
		return err
	}
	if s.Spec.Archived != false {
		return errors.New("service is not archived")
	}
	return nil
}
