.EXPORT_ALL_VARIABLES:

VERSION ?= 0.0.10

IMG_REPO = quay.io/kuberlogic
IMG_LATEST_TAG=latest

IMAGE_NAME = chargebee-integration
IMAGE_PATH = $(IMG_REPO)/$(IMAGE_NAME)

CHARGEBEE_SITE =
CHARGEBEE_KEY =
KUBERLOGIC_APISERVER_SCHEME = http
KUBERLOGIC_APISERVER_HOST = localhost:8001
KUBERLOGIC_APISERVER_TOKEN = 8ZTjsD3t2Q3Yq-C4-hoahcFn
KUBERLOGIC_DOMAIN = example.com
KUBERLOGIC_TYPE = docker-compose

SENTRY_DSN =

.PHONY: run
run:
	go run main.go

.PHONY: test
test:
	go test -v ./... -coverprofile cover.out

.PHONY: docker-build
docker-build:
	CGO_ENABLED=0 \
	GOOS=linux \
	GOARCH=amd64 \
	GO111MODULE=on \
	go build -a -o bin/chargebee-integration main.go
	docker build . \
		--build-arg BIN=bin/chargebee-integration \
		-t $(IMAGE_PATH):$(VERSION) \
		-t $(IMAGE_PATH):$(IMG_LATEST_TAG)

.PHONY: docker-push
docker-push:
	docker push $(IMAGE_PATH):$(VERSION)
	docker push $(IMAGE_PATH):$(IMG_LATEST_TAG)