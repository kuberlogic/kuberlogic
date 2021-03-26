module github.com/kuberlogic/operator/modules/watcher

go 1.13

require (
	github.com/go-sql-driver/mysql v1.5.0
	github.com/jackc/pgx/v4 v4.10.1
	github.com/kuberlogic/operator/modules/operator v0.0.21-0.20210326133005-d49b714c8e5a
	github.com/pkg/errors v0.9.1
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
	sigs.k8s.io/controller-runtime v0.8.3
)
