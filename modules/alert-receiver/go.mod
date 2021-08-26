module github.com/kuberlogic/operator/modules/alert-receiver

go 1.16

replace github.com/kuberlogic/operator/modules/operator => ../operator/

require (
	github.com/kuberlogic/operator/modules/operator v0.0.21-0.20210709150852-c26569dcc3c3
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
)
