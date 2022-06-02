module github.com/kuberlogic/kuberlogic/modules/updater

go 1.16

replace github.com/kuberlogic/kuberlogic/modules/operator => ../operator/

replace github.com/bitpoke/mysql-operator v0.5.1 => github.com/ynnt/mysql-operator v0.4.1-0.20211105080955-6cd163c75b57

require (
	github.com/coreos/go-semver v0.3.0
	github.com/kuberlogic/kuberlogic/modules/operator v0.0.21-0.20210723121420-ca52ca2c92ab
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
)
