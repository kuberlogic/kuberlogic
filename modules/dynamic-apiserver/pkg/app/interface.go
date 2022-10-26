package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiBackup "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/backup"
	apiRestore "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/restore"
	apiService "github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/restapi/operations/service"
)

type Handlers interface {
	OnShutdown()
	ListOptionsByKeyValue(key string, value *string) v1.ListOptions

	BackupAddHandler(params apiBackup.BackupAddParams, _ *models.Principal) middleware.Responder
	BackupDeleteHandler(params apiBackup.BackupDeleteParams, _ *models.Principal) middleware.Responder
	BackupListHandler(params apiBackup.BackupListParams, _ *models.Principal) middleware.Responder
	RestoreAddHandler(params apiRestore.RestoreAddParams, _ *models.Principal) middleware.Responder
	RestoreDeleteHandler(params apiRestore.RestoreDeleteParams, _ *models.Principal) middleware.Responder
	RestoreListHandler(params apiRestore.RestoreListParams, _ *models.Principal) middleware.Responder
	ServiceAddHandler(params apiService.ServiceAddParams, _ *models.Principal) middleware.Responder
	ServiceArchiveHandler(params apiService.ServiceArchiveParams, _ *models.Principal) middleware.Responder
	ServiceCredentialsUpdateHandler(params apiService.ServiceCredentialsUpdateParams, _ *models.Principal) middleware.Responder
	ServiceDeleteHandler(params apiService.ServiceDeleteParams, _ *models.Principal) middleware.Responder
	ServiceEditHandler(params apiService.ServiceEditParams, _ *models.Principal) middleware.Responder
	ServiceExplainHandler(params apiService.ServiceExplainParams, _ *models.Principal) middleware.Responder
	ServiceGetHandler(params apiService.ServiceGetParams, _ *models.Principal) middleware.Responder
	ServiceListHandler(params apiService.ServiceListParams, _ *models.Principal) middleware.Responder
	ServiceLogsHandler(params apiService.ServiceLogsParams, _ *models.Principal) middleware.Responder
	ServiceSecretsListHandler(params apiService.ServiceSecretsListParams, _ *models.Principal) middleware.Responder
	ServiceUnarchiveHandler(params apiService.ServiceUnarchiveParams, _ *models.Principal) middleware.Responder
}
