// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/auth"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
)

//go:generate swagger generate server --target ../../generated --name Kuberlogic --spec ../../../openapi.yaml --template-dir swagger-templates/templates --principal models.Principal

func configureFlags(api *operations.KuberlogicAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.KuberlogicAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.UseSwaggerUI()
	// To continue using redoc as your UI, uncomment the following line
	// api.UseRedoc()

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	// Applies when the "Authorization" header is set
	if api.BearerAuth == nil {
		api.BearerAuth = func(token string) (*models.Principal, error) {
			return nil, errors.NotImplemented("api key auth (Bearer) Authorization from header param [Authorization] has not yet been implemented")
		}
	}

	// Set your custom authorizer if needed. Default one is security.Authorized()
	// Expected interface runtime.Authorizer
	//
	// Example:
	// api.APIAuthorizer = security.Authorized()
	if api.ServiceBackupConfigCreateHandler == nil {
		api.ServiceBackupConfigCreateHandler = service.BackupConfigCreateHandlerFunc(func(params service.BackupConfigCreateParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation service.BackupConfigCreate has not yet been implemented")
		})
	}
	if api.ServiceBackupConfigDeleteHandler == nil {
		api.ServiceBackupConfigDeleteHandler = service.BackupConfigDeleteHandlerFunc(func(params service.BackupConfigDeleteParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation service.BackupConfigDelete has not yet been implemented")
		})
	}
	if api.ServiceBackupConfigEditHandler == nil {
		api.ServiceBackupConfigEditHandler = service.BackupConfigEditHandlerFunc(func(params service.BackupConfigEditParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation service.BackupConfigEdit has not yet been implemented")
		})
	}
	if api.ServiceBackupConfigGetHandler == nil {
		api.ServiceBackupConfigGetHandler = service.BackupConfigGetHandlerFunc(func(params service.BackupConfigGetParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation service.BackupConfigGet has not yet been implemented")
		})
	}
	if api.ServiceBackupListHandler == nil {
		api.ServiceBackupListHandler = service.BackupListHandlerFunc(func(params service.BackupListParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation service.BackupList has not yet been implemented")
		})
	}
	if api.ServiceDatabaseCreateHandler == nil {
		api.ServiceDatabaseCreateHandler = service.DatabaseCreateHandlerFunc(func(params service.DatabaseCreateParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation service.DatabaseCreate has not yet been implemented")
		})
	}
	if api.ServiceDatabaseDeleteHandler == nil {
		api.ServiceDatabaseDeleteHandler = service.DatabaseDeleteHandlerFunc(func(params service.DatabaseDeleteParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation service.DatabaseDelete has not yet been implemented")
		})
	}
	if api.ServiceDatabaseListHandler == nil {
		api.ServiceDatabaseListHandler = service.DatabaseListHandlerFunc(func(params service.DatabaseListParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation service.DatabaseList has not yet been implemented")
		})
	}
	if api.ServiceDatabaseRestoreHandler == nil {
		api.ServiceDatabaseRestoreHandler = service.DatabaseRestoreHandlerFunc(func(params service.DatabaseRestoreParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation service.DatabaseRestore has not yet been implemented")
		})
	}
	if api.AuthLoginUserHandler == nil {
		api.AuthLoginUserHandler = auth.LoginUserHandlerFunc(func(params auth.LoginUserParams) middleware.Responder {
			return middleware.NotImplemented("operation auth.LoginUser has not yet been implemented")
		})
	}
	if api.ServiceLogsGetHandler == nil {
		api.ServiceLogsGetHandler = service.LogsGetHandlerFunc(func(params service.LogsGetParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation service.LogsGet has not yet been implemented")
		})
	}
	if api.ServiceRestoreListHandler == nil {
		api.ServiceRestoreListHandler = service.RestoreListHandlerFunc(func(params service.RestoreListParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation service.RestoreList has not yet been implemented")
		})
	}
	if api.ServiceServiceAddHandler == nil {
		api.ServiceServiceAddHandler = service.ServiceAddHandlerFunc(func(params service.ServiceAddParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation service.ServiceAdd has not yet been implemented")
		})
	}
	if api.ServiceServiceDeleteHandler == nil {
		api.ServiceServiceDeleteHandler = service.ServiceDeleteHandlerFunc(func(params service.ServiceDeleteParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation service.ServiceDelete has not yet been implemented")
		})
	}
	if api.ServiceServiceEditHandler == nil {
		api.ServiceServiceEditHandler = service.ServiceEditHandlerFunc(func(params service.ServiceEditParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation service.ServiceEdit has not yet been implemented")
		})
	}
	if api.ServiceServiceGetHandler == nil {
		api.ServiceServiceGetHandler = service.ServiceGetHandlerFunc(func(params service.ServiceGetParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation service.ServiceGet has not yet been implemented")
		})
	}
	if api.ServiceServiceListHandler == nil {
		api.ServiceServiceListHandler = service.ServiceListHandlerFunc(func(params service.ServiceListParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation service.ServiceList has not yet been implemented")
		})
	}
	if api.ServiceUserCreateHandler == nil {
		api.ServiceUserCreateHandler = service.UserCreateHandlerFunc(func(params service.UserCreateParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation service.UserCreate has not yet been implemented")
		})
	}
	if api.ServiceUserDeleteHandler == nil {
		api.ServiceUserDeleteHandler = service.UserDeleteHandlerFunc(func(params service.UserDeleteParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation service.UserDelete has not yet been implemented")
		})
	}
	if api.ServiceUserEditHandler == nil {
		api.ServiceUserEditHandler = service.UserEditHandlerFunc(func(params service.UserEditParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation service.UserEdit has not yet been implemented")
		})
	}
	if api.ServiceUserListHandler == nil {
		api.ServiceUserListHandler = service.UserListHandlerFunc(func(params service.UserListParams, principal *models.Principal) middleware.Responder {
			return middleware.NotImplemented("operation service.UserList has not yet been implemented")
		})
	}

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
