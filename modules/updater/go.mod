module github.com/kuberlogic/operator/modules/updater

go 1.13

require (
	github.com/coreos/go-semver v0.3.0
	github.com/kuberlogic/operator/modules/operator v0.0.21-0.20210324085705-c670b6a252cb
	//github.com/kuberlogic/operator/modules/operator v0.0.20-0.20210317122412-00275b30510c
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.0+incompatible
)

// Pin k8s deps to 1.18.8
replace k8s.io/client-go => k8s.io/client-go v0.18.8
