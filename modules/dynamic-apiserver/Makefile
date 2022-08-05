.EXPORT_ALL_VARIABLES:

VERSION ?= 0.0.11

IMG_REPO = quay.io/kuberlogic
IMG_LATEST_TAG=latest

APISERVER_NAME = dynamic-apiserver
APISERVER_IMG = $(IMG_REPO)/$(APISERVER_NAME)

KUBERLOGIC_APISERVER_TOKEN=8ZTjsD3t2Q3Yq-C4-hoahcFn
KUBERLOGIC_SENTRY_DSN =
KUBERLOGIC_DEBUG_LOGS = true

.PHONY: recreate-generated-server
recreate-generated-server:
	rm -rf pkg/generated/models pkg/generated/restapi
	mkdir -p pkg/generated

.PHONY: recreate-generated-client
recreate-generated-client:
	rm -rf pkg/generated/client
	mkdir -p pkg/generated

.PHONY: generate-server
generate-server: recreate-generated-server
	swagger generate server \
		--spec openapi.yaml \
		--target pkg/generated \
		-P models.Principal \
		--config-file swagger-templates/default-server.yml \
		--template-dir swagger-templates/templates/ \
		--name kuberlogic

.PHONY: generate-client
generate-client: recreate-generated-client
	swagger generate client \
		--spec openapi.yaml \
		--target pkg/generated/

.PHONY: test
test: cli-configs
	go test -v ./... -coverprofile cover.out

.PHONY: cli-configs
cli-configs:
	cd ../dynamic-operator && zip -r kustomize-configs.zip config/ && mv kustomize-configs.zip ../dynamic-apiserver/pkg/cli/


.PHONY: build-cli
build-cli: cli-configs
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 GO111MODULE=on go build -o kuberlogic_x86_64_darwin cmd/cli/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 GO111MODULE=on go build -o kuberlogic_arm64_darwin cmd/cli/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o kuberlogic_x86_64_linux cmd/cli/main.go
	test -e kuberlogic && rm kuberlogic || echo "file does not exist"
	cp kuberlogic_$(shell uname -m)_$(shell uname | tr '[:upper:]' '[:lower:]') kuberlogic

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
