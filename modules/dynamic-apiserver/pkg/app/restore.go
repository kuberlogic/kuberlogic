/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package app

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/api"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/logging"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

type ExtendedRestoreGetter interface {
	Restores() ExtendedRestoreInterface
}

type ExtendedRestoreInterface interface {
	api.RestoreInterface
	ListByServiceName(ctx context.Context, service *string) (*v1alpha1.KuberlogicServiceRestoreList, error)
	Wait(ctx context.Context, log logging.Logger, restoreId string, maxRetries int) error
}

type restores struct {
	api.RestoreInterface
}

var _ ExtendedRestoreInterface = &restores{}

func newRestores(c rest.Interface) ExtendedRestoreInterface {
	s := &restores{}
	s.RestoreInterface = api.NewRestores(c)
	return s
}

func (r *restores) ListByServiceName(ctx context.Context, service *string) (*v1alpha1.KuberlogicServiceRestoreList, error) {
	opts := v1.ListOptions{}
	if service != nil {
		labelSelector := v1.LabelSelector{
			MatchLabels: map[string]string{util.BackupRestoreServiceField: *service},
		}
		opts = v1.ListOptions{
			LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		}
	}
	return r.List(ctx, opts)
}

func (r *restores) Wait(ctx context.Context, log logging.Logger, restoreId string, maxRetries int) error {
	timeout := time.Second
	for i := maxRetries; i > 0; i-- {
		result, err := r.Get(ctx, restoreId, v1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "Error while getting service restore")
		}

		if ok := result.IsSuccessful(); ok {
			return nil
		}

		timeout = timeout * 2
		time.Sleep(timeout)
		log.Infof("restore is not successful, trying after %s. Left %d retries", timeout, i-1)
	}
	return errors.New("Retries exceeded, restore is not successful")
}
