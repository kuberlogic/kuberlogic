.EXPORT_ALL_VARIABLES:

KUBERLOGIC_AUTH_PROVIDER = none

#  Current apiserver version
VERSION ?= 0.0.24
# private repo for images
IMG_REPO = quay.io/kuberlogic
NAME = apiserver
# Image URL to use all building/pushing image targets
IMG ?= $(IMG_REPO)/$(NAME):$(VERSION)

KUBERLOGIC_DEBUGLOGS=True
SENTRY_DSN =

clean:
	@rm -rf ./cmd internal/generated  internal/config
	@mkdir -p cmd
	@mkdir -p internal/generated
	@mkdir -p internal/app
	@mkdir -p internal/config


gen: clean
	@swagger generate server \
		-f ./openapi.yaml \
		-t ./internal/generated \
		-C ./swagger-templates/default-server.yml \
		-P models.Principal \
		--template-dir ./swagger-templates/templates \
		--name kuberlogic

fmt:
	go fmt ./...

vet:
	go vet ./...

test: fmt vet
	#go test ./... -coverprofile cover.out

run: test
	@go run -race main.go --scheme=http

build: gen test
	echo "Building apiserver image"
	docker build . -t ${IMG}

push:
	docker push ${IMG}