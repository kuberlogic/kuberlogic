module github.com/kuberlogic/kuberlogic/modules/updater

go 1.16

replace github.com/kuberlogic/kuberlogic/modules/operator => ../operator/

require (
	github.com/coreos/go-semver v0.3.0
	github.com/kuberlogic/kuberlogic/modules/operator v0.0.21-0.20210723121420-ca52ca2c92ab
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
)
