.EXPORT_ALL_VARIABLES:

VERSION ?= 0.0.6

IMG_REPO = quay.io/kuberlogic
IMG_LATEST_TAG=latest

APISERVER_NAME = dynamic-apiserver
APISERVER_IMG = $(IMG_REPO)/$(APISERVER_NAME)

KUBERLOGIC_APISERVER_TOKEN=8ZTjsD3t2Q3Yq-C4-hoahcFn

.PHONY: recreate-generated-server
recreate-generated-server:
	rm -r internal/generated/models internal/generated/restapi
	mkdir -p internal/generated

.PHONY: recreate-generated-client
recreate-generated-client:
	rm -r internal/generated/client
	mkdir -p internal/generated

.PHONY: generate-server
generate-server: recreate-generated-server
	swagger generate server \
		--spec openapi.yaml \
		--target internal/generated \
		-P models.Principal \
		--config-file swagger-templates/default-server.yml \
		--template-dir swagger-templates/templates/ \
		--name kuberlogic

.PHONY: generate-client
generate-client: recreate-generated-client
	swagger generate client \
		--spec openapi.yaml \
		--target internal/generated/

.PHONY: test
test:
	go test -v ./... -coverprofile cover.out

.PHONY: build-cli
build-cli:
	go build -o kuberlogic cmd/cli/main.go

.PHONY: run
run:
	go run main.go

.PHONY: generate
generate: generate-server generate-client

.PHONY: docker-build
docker-build:
	CGO_ENABLED=0 \
	GOOS=linux \
	GOARCH=amd64 \
	GO111MODULE=on \
	go build -a -o bin/apiserver main.go
	docker build . \
		--build-arg BIN=bin/apiserver \
		-t $(APISERVER_IMG):$(VERSION) \
		-t $(APISERVER_IMG):$(IMG_LATEST_TAG)

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push $(APISERVER_IMG):$(VERSION)
	docker push $(APISERVER_IMG):$(IMG_LATEST_TAG)