module github.com/kuberlogic/operator/alert-receiver

go 1.13

require (
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.0+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.18.8
