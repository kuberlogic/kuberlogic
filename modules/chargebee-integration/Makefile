.EXPORT_ALL_VARIABLES:

VERSION = 0.0.13

IMG_REPO = quay.io/kuberlogic
IMG_LATEST_TAG=latest

IMAGE_NAME = chargebee-integration
IMAGE_PATH = $(IMG_REPO)/$(IMAGE_NAME)

CHARGEBEE_SITE ?=
CHARGEBEE_KEY ?=
KUBERLOGIC_APISERVER_SCHEME = http
KUBERLOGIC_APISERVER_HOST = localhost:8001
KUBERLOGIC_APISERVER_TOKEN = 8ZTjsD3t2Q3Yq-C4-hoahcFn
KUBERLOGIC_TYPE = docker-compose

SENTRY_DSN =

ifeq ($(DEV_BUILD),true)
	VERSION := $(VERSION)-$(shell git rev-list --count $(shell git rev-parse --abbrev-ref HEAD))
endif
COMMIT_SHA = $(shell git rev-parse HEAD)
IMG_SHA_TAG ?= $(VERSION)-$(COMMIT_SHA)

.PHONY: run
run:
	go run main.go ${ARGS}

.PHONY: test
test:
	go test -v ./... -coverprofile cover.out

.PHONY: docker-build
docker-build: export LDFLAGS="-X main.ver=$(VERSION)"
docker-build:
	CGO_ENABLED=0 \
	GOOS=linux \
	GOARCH=amd64 \
	GO111MODULE=on \
	go build -a -o bin/chargebee-integration -ldflags=$(LDFLAGS) main.go
	docker build . \
		--build-arg BIN=bin/chargebee-integration \
		-t $(IMAGE_PATH):$(VERSION) \
		-t $(IMAGE_PATH):$(IMG_LATEST_TAG) \
		-t $(IMAGE_PATH):$(IMG_SHA_TAG)

.PHONY: docker-push
docker-push:
	docker push $(IMAGE_PATH):$(VERSION)
	docker push $(IMAGE_PATH):$(IMG_LATEST_TAG)

.PHONY: docker-push-cache
docker-push-cache:
	docker push $(IMAGE_PATH):$(IMG_SHA_TAG)

.PHONY: docker-pull-cache
docker-pull-cache:
	docker pull $(IMAGE_PATH):$(IMG_SHA_TAG)

.PHONY: docker-restore-cache
docker-restore-cache: docker-pull-cache
	docker tag $(IMAGE_PATH):$(IMG_SHA_TAG) $(IMAGE_PATH):$(VERSION)
	docker tag $(IMAGE_PATH):$(IMG_SHA_TAG) $(IMAGE_PATH):$(IMG_LATEST_TAG)
