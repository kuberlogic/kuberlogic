module github.com/kuberlogic/operator/modules/updater

go 1.16

replace (
	github.com/kuberlogic/operator/modules/operator => ../operator/
)

require (
	github.com/coreos/go-semver v0.3.0
	github.com/kuberlogic/operator/modules/operator v0.0.21-0.20210723121420-ca52ca2c92ab // indirect
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
)
