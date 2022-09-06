module github.com/kuberlogic/kuberlogic/modules/dynamic-operator

go 1.16

require (
	github.com/compose-spec/compose-go v1.2.4
	github.com/getsentry/sentry-go v0.13.0
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/go-test/deep v1.0.8
	github.com/hashicorp/go-hclog v1.1.0
	github.com/hashicorp/go-plugin v1.4.3
	github.com/kuberlogic/zapsentry v1.6.2
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.16.0
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron v1.1.0
	github.com/vmware-tanzu/velero v1.8.1
	github.com/vrischmann/envconfig v1.3.0
	github.com/zalando/postgres-operator v1.7.1
	go.uber.org/zap v1.19.0
	k8s.io/api v0.22.3
	k8s.io/apiextensions-apiserver v0.22.3
	k8s.io/apimachinery v0.22.3
	k8s.io/client-go v0.22.3
	k8s.io/utils v0.0.0-20210930125809-cb0fa318a74b
	sigs.k8s.io/controller-runtime v0.10.2
)
