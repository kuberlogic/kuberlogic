/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package app

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/api"
)

type ExtendedServiceGetter interface {
	Services() ExtendedServiceInterface
}

type ExtendedServiceInterface interface {
	api.ServiceInterface
	Exists(ctx context.Context, opts v1.ListOptions) (bool, error)
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

func (svc *services) Exists(ctx context.Context, opts v1.ListOptions) (bool, error) {
	r, err := svc.List(ctx, opts)
	if err != nil {
		return false, err
	}
	return len(r.Items) > 0, nil
}
