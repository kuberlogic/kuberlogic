module github.com/kuberlogic/operator/modules/apiserver

go 1.13

require (
	github.com/aws/aws-sdk-go v1.36.29
	github.com/casbin/casbin/v2 v2.19.4
	github.com/coreos/go-oidc v2.1.0+incompatible
	github.com/dgraph-io/ristretto v0.0.3
	github.com/getsentry/sentry-go v0.10.0
	github.com/go-chi/chi v1.5.1
	github.com/go-openapi/errors v0.19.9
	github.com/go-openapi/loads v0.19.7
	github.com/go-openapi/runtime v0.19.24
	github.com/go-openapi/spec v0.19.15
	github.com/go-openapi/strfmt v0.19.11
	github.com/go-openapi/swag v0.19.12
	github.com/go-openapi/validate v0.19.15
	github.com/jessevdk/go-flags v1.4.0
	github.com/kuberlogic/operator/modules/operator v0.0.21-0.20210326133005-d49b714c8e5a
	github.com/kuberlogic/zapsentry v1.6.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.10.0
	github.com/vrischmann/envconfig v1.3.0
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
)
