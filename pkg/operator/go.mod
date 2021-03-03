module github.com/kuberlogic/operator/pkg/operator

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/pkg/errors v0.9.1
	github.com/presslabs/mysql-operator v0.4.0
	github.com/prometheus/client_golang v1.0.0
	github.com/spotahome/redis-operator v1.0.0
	github.com/zalando/postgres-operator v1.5.1-0.20200903060246-03437b63749e
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.2
)

replace (
	// using forks for [postgres/mysql/redis]-operator (the same api version)
	github.com/presslabs/mysql-operator => github.com/cloudlinux/mysql-operator v0.4.1-0.20200922131437-71ac68b234d0
	github.com/spotahome/redis-operator => github.com/cloudlinux/redis-operator v1.0.1-0.20200922144448-ea17b0f10a01
	github.com/zalando/postgres-operator => github.com/cloudlinux/postgres-operator v1.5.1-0.20200922100439-a33d339eac3f
	// Pin k8s deps to 1.18.8
	k8s.io/client-go => k8s.io/client-go v0.18.8
)
