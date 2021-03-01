module github.com/kuberlogic/operator/alert-receiver

go 1.13

require (
	github.com/kuberlogic/operator v0.0.19-0.20210219103633-842d7ac86fe2
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.0+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.18.8
