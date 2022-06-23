module chargebee_integration

go 1.16

replace github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver => ../dynamic-apiserver/

require (
	github.com/chargebee/chargebee-go v2.10.0+incompatible
	github.com/go-openapi/runtime v0.23.3
	github.com/go-openapi/strfmt v0.21.2
	github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver v0.0.0-20220329063704-75e3ccc06da7
	github.com/spf13/viper v1.7.0
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.21.0
)
