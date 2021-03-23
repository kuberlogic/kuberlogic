module github.com/kuberlogic/operator/modules/alert-receiver

go 1.13

require (
	github.com/TheZeroSlave/zapsentry v1.6.0 // indirect
	github.com/kuberlogic/operator/modules/operator v0.0.20-0.20210323085615-3c2beb638ed9 // indirect
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.0+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.18.8
