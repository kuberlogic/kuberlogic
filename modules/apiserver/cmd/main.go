package cmd

import (
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-openapi/loads"
	"github.com/jessevdk/go-flags"

	"github.com/kuberlogic/operator/modules/apiserver/internal/app"
	"github.com/kuberlogic/operator/modules/apiserver/internal/cache"
	"github.com/kuberlogic/operator/modules/apiserver/internal/config"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations"

	apiAuth "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/auth"

	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/internal/logging"
	apiserverMiddleware "github.com/kuberlogic/operator/modules/apiserver/internal/net/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/security"
	"github.com/kuberlogic/operator/modules/apiserver/util/k8s"
	cloudlinuxv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/util"
	"k8s.io/client-go/kubernetes"
	k8scheme "k8s.io/client-go/kubernetes/scheme"
)

func Main(args []string) {
	mainLog := logging.WithComponentLogger("main")
	cfg, err := config.InitConfig("kuberlogic", logging.WithComponentLogger("config"))
	if err != nil {
		mainLog.Fatalw("", "error", err)
		os.Exit(1)
	}
	logging.DebugLevel(cfg.DebugLogs)

	// init sentry
	if dsn := cfg.Sentry.Dsn; dsn != "" {
		logging.UseSentry(dsn)

		mainLog.Debugw("sentry for apiserver was initialized")

		// Flush buffered events before the program terminates.
		defer sentry.Flush(2 * time.Second)
	}

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		mainLog.Fatalw("swagger does not loaded", "error", err)
	}

	cache_, err := cache.NewCache(logging.WithComponentLogger("cache"))
	if err != nil {
		mainLog.Fatalw("not initialized cache", "error", err)
	}

	authProvider, err := security.NewAuthProvider(cfg, cache_, logging.WithComponentLogger("auth"))
	if err != nil {
		mainLog.Fatalw("not initialized auth provider", "error", err)
	}

	k8sconf, err := k8s.GetConfig(cfg)
	if err != nil {
		mainLog.Fatalw("could not get config", "error", err)
	}

	err = cloudlinuxv1.AddToScheme(k8scheme.Scheme)
	if err != nil {
		mainLog.Fatalw("could not add to scheme", "error", err)
	}

	crdClient, err := util.GetKuberLogicClient(k8sconf)
	if err != nil {
		mainLog.Fatalw("could not generate rest client", "error", err)
	}

	baseClient, err := kubernetes.NewForConfig(k8sconf)
	if err != nil {
		mainLog.Fatalw("could not get base client", "error", err)
	}

	srv := app.New(baseClient, crdClient, authProvider, logging.WithComponentLogger("server"))
	api := operations.NewKuberlogicAPI(swaggerSpec)

	api.ServiceBackupConfigCreateHandler = apiService.BackupConfigCreateHandlerFunc(srv.BackupConfigCreateHandler)
	api.ServiceBackupConfigDeleteHandler = apiService.BackupConfigDeleteHandlerFunc(srv.BackupConfigDeleteHandler)
	api.ServiceBackupConfigEditHandler = apiService.BackupConfigEditHandlerFunc(srv.BackupConfigEditHandler)
	api.ServiceBackupConfigGetHandler = apiService.BackupConfigGetHandlerFunc(srv.BackupConfigGetHandler)
	api.ServiceBackupListHandler = apiService.BackupListHandlerFunc(srv.BackupListHandler)
	api.ServiceDatabaseCreateHandler = apiService.DatabaseCreateHandlerFunc(srv.DatabaseCreateHandler)
	api.ServiceDatabaseDeleteHandler = apiService.DatabaseDeleteHandlerFunc(srv.DatabaseDeleteHandler)
	api.ServiceDatabaseListHandler = apiService.DatabaseListHandlerFunc(srv.DatabaseListHandler)
	api.ServiceDatabaseRestoreHandler = apiService.DatabaseRestoreHandlerFunc(srv.DatabaseRestoreHandler)
	api.AuthLoginUserHandler = apiAuth.LoginUserHandlerFunc(srv.LoginUserHandler)
	api.ServiceLogsGetHandler = apiService.LogsGetHandlerFunc(srv.LogsGetHandler)
	api.ServiceRestoreListHandler = apiService.RestoreListHandlerFunc(srv.RestoreListHandler)
	api.ServiceServiceAddHandler = apiService.ServiceAddHandlerFunc(srv.ServiceAddHandler)
	api.ServiceServiceDeleteHandler = apiService.ServiceDeleteHandlerFunc(srv.ServiceDeleteHandler)
	api.ServiceServiceEditHandler = apiService.ServiceEditHandlerFunc(srv.ServiceEditHandler)
	api.ServiceServiceGetHandler = apiService.ServiceGetHandlerFunc(srv.ServiceGetHandler)
	api.ServiceServiceListHandler = apiService.ServiceListHandlerFunc(srv.ServiceListHandler)
	api.ServiceUserCreateHandler = apiService.UserCreateHandlerFunc(srv.UserCreateHandler)
	api.ServiceUserDeleteHandler = apiService.UserDeleteHandlerFunc(srv.UserDeleteHandler)
	api.ServiceUserEditHandler = apiService.UserEditHandlerFunc(srv.UserEditHandler)
	api.ServiceUserListHandler = apiService.UserListHandlerFunc(srv.UserListHandler)
	api.BearerAuth = srv.BearerAuthentication
	api.Logger = logging.WithComponentLogger("api").Infow
	api.ServerShutdown = srv.OnShutdown
	server := restapi.NewServer(api)
	defer server.Shutdown()

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "KuberLogic API"
	parser.LongDescription = "This is a KuberLogic API"
	server.ConfigureFlags()
	for _, optsGroup := range api.CommandLineOptionsGroups {
		_, err := parser.AddGroup(optsGroup.ShortDescription, optsGroup.LongDescription, optsGroup.Options)
		if err != nil {
			mainLog.Fatalw("could not add group", "error", err)
		}
	}

	if _, err := parser.ParseArgs(args); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}

	h := api.Serve(nil)
	r := chi.NewRouter()
	r.Use(apiserverMiddleware.NewLoggingMiddleware(logging.WithComponentLogger("request-handler")))
	r.Use(middleware.Recoverer)
	r.Use(apiserverMiddleware.SentryLogPanic)
	r.Use(apiserverMiddleware.SetSentryRequestScope)
	r.Mount("/", h)

	server.ConfigureAPI()
	server.SetHandler(r)

	server.Port = cfg.HTTPBindPort
	server.Host = cfg.BindHost
	if err := server.Serve(); err != nil {
		mainLog.Fatalw("problem with serve server", "error", err)
	}
}
