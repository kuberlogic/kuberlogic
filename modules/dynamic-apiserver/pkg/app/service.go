/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package app

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/api"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

type ExtendedServiceGetter interface {
	Services() ExtendedServiceInterface
}

type ExtendedServiceInterface interface {
	api.ServiceInterface
	ListByFieldLabel(ctx context.Context, field string, value *string) (*v1alpha1.KuberLogicServiceList, error)
	IsSubscriptionAlreadyExist(ctx context.Context, subscriptionId *string) (bool, error)
}

type services struct {
	api.ServiceInterface
}

var _ ExtendedServiceInterface = &services{}

func newServices(c rest.Interface) ExtendedServiceInterface {
	s := &services{}
	s.ServiceInterface = api.NewServices(c)
	return s
}

func (svc *services) ListByFieldLabel(ctx context.Context, field string, value *string) (*v1alpha1.KuberLogicServiceList, error) {
	opts := v1.ListOptions{}
	if value != nil {
		labelSelector := v1.LabelSelector{
			MatchLabels: map[string]string{field: *value},
		}
		opts = v1.ListOptions{
			LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		}
	}
	return svc.List(ctx, opts)
}

func (svc *services) IsSubscriptionAlreadyExist(ctx context.Context, subscriptionId *string) (bool, error) {
	r, err := svc.ListByFieldLabel(ctx, util.SubscriptionField, subscriptionId)
	if err != nil {
		return false, err
	}
	return len(r.Items) > 0, nil
}
