module github.com/kuberlogic/operator/modules/alert-receiver

go 1.13

require (
	github.com/kuberlogic/operator/modules/operator v0.0.0-20210305115707-d1e7df28ca4a
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.0+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.18.8
