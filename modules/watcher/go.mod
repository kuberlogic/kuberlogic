module github.com/kuberlogic/operator/modules/watcher

go 1.13

require (
	github.com/go-sql-driver/mysql v1.5.0
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/jackc/pgx/v4 v4.10.1
	github.com/kuberlogic/operator/modules/operator v0.0.0-20210315110409-983436e5ed87
	github.com/pkg/errors v0.9.1
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20200815180417-3bc9d57fc792 // indirect
	sigs.k8s.io/controller-runtime v0.6.2
)

// Pin k8s deps to 0.18.8
replace (
	k8s.io/api => k8s.io/api v0.18.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.8
	k8s.io/client-go => k8s.io/client-go v0.18.8
)
