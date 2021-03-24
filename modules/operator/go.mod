module github.com/kuberlogic/operator/modules/operator

go 1.16

require (
	github.com/go-logr/logr v0.4.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/jackc/pgx/v4 v4.10.1
	github.com/pkg/errors v0.9.1
	github.com/presslabs/mysql-operator v0.5.0-rc.2
	github.com/prometheus/client_golang v1.7.1
	github.com/zalando/postgres-operator v1.6.1
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
	sigs.k8s.io/controller-runtime v0.8.3
)
