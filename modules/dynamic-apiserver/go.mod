module github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver

go 1.16

replace github.com/kuberlogic/kuberlogic/modules/dynamic-operator => ../dynamic-operator/

require (
	github.com/AlecAivazis/survey/v2 v2.3.5
	github.com/getsentry/sentry-go v0.13.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-chi/chi v1.5.4
	github.com/go-chi/cors v1.2.0
	github.com/go-openapi/errors v0.20.2
	github.com/go-openapi/loads v0.21.1
	github.com/go-openapi/runtime v0.23.3
	github.com/go-openapi/spec v0.20.4
	github.com/go-openapi/strfmt v0.21.2
	github.com/go-openapi/swag v0.21.1
	github.com/go-openapi/validate v0.21.0
	github.com/google/uuid v1.2.0
	github.com/jessevdk/go-flags v1.5.0
	github.com/kuberlogic/kuberlogic/modules/dynamic-operator v0.0.0-20220329063704-75e3ccc06da7
	github.com/novln/docker-parser v1.0.0
	github.com/olekukonko/tablewriter v0.0.5
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.5.0
	github.com/spf13/viper v1.12.0
	github.com/vrischmann/envconfig v1.3.0
	github.com/xeipuuv/gojsonschema v1.2.0
	go.uber.org/zap v1.21.0
	golang.org/x/net v0.0.0-20220520000938-2e3eb7b945c2
	k8s.io/api v0.22.3
	k8s.io/apimachinery v0.22.3
	k8s.io/client-go v0.22.3
	k8s.io/utils v0.0.0-20210930125809-cb0fa318a74b
)
