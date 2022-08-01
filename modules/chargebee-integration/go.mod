module chargebee_integration

go 1.16

replace (
	github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver => ../dynamic-apiserver/
	github.com/kuberlogic/kuberlogic/modules/dynamic-operator => ../dynamic-operator/
)

require (
	github.com/chargebee/chargebee-go v2.10.0+incompatible
	github.com/dustinkirkland/golang-petname v0.0.0-20191129215211-8e5a1ed0cff0
	github.com/getsentry/sentry-go v0.13.0
	github.com/go-openapi/runtime v0.23.3
	github.com/go-openapi/strfmt v0.21.2
	github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver v0.0.0-20220329063704-75e3ccc06da7
	github.com/kuberlogic/kuberlogic/modules/dynamic-operator v0.0.0-20220329063704-75e3ccc06da7
	github.com/spf13/viper v1.9.0
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.21.0
)
