.EXPORT_ALL_VARIABLES:

KUBERLOGIC_AUTH_PROVIDER = none

#  Current apiserver version
VERSION ?= 0.0.26
# private repo for images
IMG_REPO = quay.io/kuberlogic
NAME = apiserver
# Image URL to use all building/pushing image targets
IMG ?= $(IMG_REPO)/$(NAME):$(VERSION)
IMG_LATEST = $(IMG_REPO)/$(NAME):latest
IMG_TESTS ?= $(IMG_REPO)/integration-tests:$(VERSION)
IMG_TESTS_LATEST ?= $(IMG_REPO)/integration-tests:latest


KUBERLOGIC_DEBUGLOGS=True
SENTRY_DSN =

clean:
	@rm -rf ./cmd internal/generated
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
	docker build . -t ${IMG} -t ${IMG_LATEST}

push:
	docker push ${IMG}
	docker push ${IMG_LATEST}

build-tests: gen test
	echo "Building tests image"
	docker build . -f Dockerfile.tests -t ${IMG_TESTS} -t ${IMG_TESTS_LATEST}

push-tests:
	docker push ${IMG_TESTS}
	docker push ${IMG_TESTS_LATEST}