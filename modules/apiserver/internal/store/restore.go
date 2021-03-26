package store

import (
	"context"
	"github.com/go-openapi/strfmt"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
)

const (
	kuberlogicRestoreK8sResource = "kuberlogicbackuprestores"
)

// returns empty list of no restores found
func (s *ServiceStore) GetServiceRestores(ns, serviceName string, ctx context.Context) ([]*models.Restore, *ServiceError) {
	k8srestores := new(kuberlogicv1.KuberLogicBackupRestoreList)
	restores := make([]*models.Restore, 0)

	// todo: add field selector here
	err := s.restClient.Get().Resource(kuberlogicRestoreK8sResource).
		Namespace(ns).
		Do(ctx).
		Into(k8srestores)
	s.log.Debugw("got restores from cluster", "object", k8srestores)
	if err != nil {
		s.log.Errorw("error getting restores", "error", err)
		return restores, &ServiceError{Err: err, ClientMsg: "error getting restores for the service"}
	}

	for _, r := range k8srestores.Items {
		if r.Spec.ClusterName != serviceName {
			continue
		}
		converted, err := kuberlogicRestoreToRestore(&r)
		if err != nil {
			s.log.Errorw("error converting kubernetes restore to models", "error", err)
			continue
		}
		restores = append(restores, converted)
	}
	return restores, nil
}

func kuberlogicRestoreToRestore(r *kuberlogicv1.KuberLogicBackupRestore) (*models.Restore, error) {
	res := new(models.Restore)

	status := r.Status.Status
	datetime, err := strfmt.ParseDateTime(r.Status.CompletionTime)
	if err != nil {
		return nil, err
	}

	res.File = &r.Spec.Backup
	res.Status = &status
	res.Time = &datetime
	res.Database = &r.Spec.Database

	return res, nil
}
