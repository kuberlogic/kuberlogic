.EXPORT_ALL_VARIABLES:

CHARGEBEE_SITE ?=
CHARGEBEE_KEY ?=
KUBERLOGIC_APISERVER_SCHEME = http
KUBERLOGIC_APISERVER_HOST = localhost:8001
KUBERLOGIC_APISERVER_TOKEN = 8ZTjsD3t2Q3Yq-C4-hoahcFn
KUBERLOGIC_TYPE = docker-compose
SENTRY_DSN =

.PHONY: run
run:
	go run main.go ${ARGS}

.PHONY: test
test:
	go test -v ./... -coverprofile cover.out

.PHONY: docker-build
build: export LDFLAGS="-X main.ver=$(VERSION)"
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on \
	go build -a -o bin/chargebee-integration -ldflags=$(LDFLAGS) main.go
