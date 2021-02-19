module gitlab.com/cloudmanaged/operator/alert-receiver

go 1.13

require (
	gitlab.com/cloudmanaged/operator v0.0.19-0.20210218130521-d455734b94b6
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.0+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.18.8
