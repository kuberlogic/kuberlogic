module github.com/kuberlogic/operator/modules/operator

go 1.16

require (
	github.com/getsentry/sentry-go v0.10.0
	github.com/go-errors/errors v1.0.1
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/google/uuid v1.1.2
	github.com/jackc/pgx/v4 v4.10.1
	github.com/kuberlogic/zapsentry v1.6.2
	github.com/pkg/errors v0.9.1
	github.com/presslabs/mysql-operator v0.5.0-rc.2
	github.com/prometheus/client_golang v1.7.1
	github.com/vrischmann/envconfig v1.3.0
	github.com/zalando/postgres-operator v1.6.2
	go.uber.org/zap v1.16.0
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
	sigs.k8s.io/controller-runtime v0.8.3
)
