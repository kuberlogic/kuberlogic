module github.com/kuberlogic/operator/updater

go 1.13

require (
	github.com/coreos/go-semver v0.3.0
	github.com/kuberlogic/operator v0.0.19-0.20210219103633-842d7ac86fe2
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.0+incompatible
)

// Pin k8s deps to 1.18.8
replace k8s.io/client-go => k8s.io/client-go v0.18.8
