module github.com/kuberlogic/kuberlogic/modules/operator

go 1.16

require (
	github.com/bitpoke/mysql-operator v0.5.1
	github.com/getsentry/sentry-go v0.10.0
	github.com/go-errors/errors v1.0.1
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/google/uuid v1.1.2
	github.com/jackc/pgx/v4 v4.10.1
	github.com/kuberlogic/zapsentry v1.6.2
	github.com/mitchellh/mapstructure v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/vrischmann/envconfig v1.3.0
	github.com/zalando/postgres-operator v1.6.2
	go.uber.org/zap v1.18.1
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	sigs.k8s.io/controller-runtime v0.9.2
)

replace github.com/bitpoke/mysql-operator v0.5.1 => github.com/ynnt/mysql-operator v0.4.1-0.20211105080955-6cd163c75b57
