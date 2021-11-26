module github.com/kuberlogic/kuberlogic/modules/alert-receiver

go 1.16

replace github.com/kuberlogic/kuberlogic/modules/operator => ../operator/

replace github.com/bitpoke/mysql-operator v0.5.1 => github.com/ynnt/mysql-operator v0.4.1-0.20211105080955-6cd163c75b57

require (
	github.com/kuberlogic/kuberlogic/modules/operator v0.0.21-0.20210709150852-c26569dcc3c3
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/vrischmann/envconfig v1.3.0
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
)
