module github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver

go 1.16

replace github.com/kuberlogic/kuberlogic/modules/dynamic-operator => ../dynamic-operator/

require (
	github.com/go-chi/chi v1.5.4
	github.com/go-chi/cors v1.2.0
	github.com/go-openapi/errors v0.20.2
	github.com/go-openapi/loads v0.21.1
	github.com/go-openapi/runtime v0.23.3
	github.com/go-openapi/spec v0.20.4
	github.com/go-openapi/strfmt v0.21.2
	github.com/go-openapi/swag v0.21.1
	github.com/go-openapi/validate v0.21.0
	github.com/jessevdk/go-flags v1.5.0
	github.com/kuberlogic/kuberlogic/modules/dynamic-operator v0.0.0-20220329063704-75e3ccc06da7
	github.com/kuberlogic/zapsentry v1.6.2
	github.com/pkg/errors v0.9.1
	github.com/vrischmann/envconfig v1.3.0
	go.uber.org/zap v1.21.0
	golang.org/x/net v0.0.0-20220403103023-749bd193bc2b
	k8s.io/api v0.22.3
	k8s.io/apimachinery v0.22.3
	k8s.io/client-go v0.22.3
)
