module github.com/kuberlogic/operator/modules/updater

go 1.13

require (
	github.com/coreos/go-semver v0.3.0
	github.com/kuberlogic/operator/modules/operator v0.0.0-20210305115707-d1e7df28ca4a
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.0+incompatible
)

// Pin k8s deps to 1.18.8
replace k8s.io/client-go => k8s.io/client-go v0.18.8
