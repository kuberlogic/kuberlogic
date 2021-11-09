module github.com/kuberlogic/kuberlogic/modules/alert-receiver

go 1.16

replace github.com/kuberlogic/kuberlogic/modules/operator => ../operator/

replace github.com/bitpoke/mysql-operator v0.5.1 => github.com/ynnt/mysql-operator v0.4.1-0.20211104191942-46957122d7a4

require (
	github.com/kuberlogic/kuberlogic/modules/operator v0.0.21-0.20210709150852-c26569dcc3c3
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
)
