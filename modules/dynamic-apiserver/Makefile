.EXPORT_ALL_VARIABLES:

KUBERLOGIC_APISERVER_TOKEN=8ZTjsD3t2Q3Yq-C4-hoahcFn
KUBERLOGIC_SENTRY_DSN =
KUBERLOGIC_DEBUG_LOGS = true
KUBERLOGIC_DOMAIN = example.com

PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))

SWAGGER = $(shell pwd)/bin/swagger
.PHONY: swagger
swagger: ## Download swagger locally if necessary.
	$(call go-get-tool,$(SWAGGER),github.com/go-swagger/go-swagger/cmd/swagger@v0.29.0)

.PHONY: recreate-generated-server
recreate-generated-server:
	rm -rf pkg/generated/models pkg/generated/restapi
	mkdir -p pkg/generated

.PHONY: recreate-generated-client
recreate-generated-client:
	rm -rf pkg/generated/client
	mkdir -p pkg/generated

.PHONY: generate-server
generate-server: recreate-generated-server swagger
	$(SWAGGER) generate server \
		--spec openapi.yaml \
		--target pkg/generated \
		-P models.Principal \
		--config-file swagger-templates/default-server.yml \
		--template-dir swagger-templates/templates/ \
		--name kuberlogic

.PHONY: generate-client
generate-client: recreate-generated-client swagger
	$(SWAGGER) generate client \
		--spec openapi.yaml \
		--target pkg/generated/

.PHONY: test
test: build-configs
	go test -v ./... -coverprofile cover.out ${ARGS}

.PHONY: build-configs
build-configs:
	test -e pkg/cli/kustomize-configs.zip && rm -f pkg/cli/kustomize-configs.zip || echo "archive does not exist"
	cd ../../ && $(MAKE) set-version
	cd ../dynamic-operator/ && zip -r ../dynamic-apiserver/pkg/cli/kustomize-configs.zip config/

.PHONY: build-cli
build-cli: export CGO_ENABLED=0
build-cli: export GO111MODULE=on
build-cli: export LDFLAGS="-X github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/cli.ver=$(VERSION)"
build-cli: build-configs
	GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS) -a -o bin/kuberlogic_x86_64_darwin cmd/cli/main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags $(LDFLAGS) -a -o bin/kuberlogic_arm64_darwin cmd/cli/main.go
	GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -a -o bin/kuberlogic_x86_64_linux cmd/cli/main.go
	test -e kuberlogic && rm kuberlogic || echo "link does not exist"
	ln -s bin/kuberlogic_$(shell uname -m)_$(shell uname | tr '[:upper:]' '[:lower:]') kuberlogic

.PHONY: build
build: export CGO_ENABLED=0
build: export GO111MODULE=on
build: export GOOS=linux
build: export GOARCH=amd64
build:
	go build -a -o bin/apiserver -ldflags=$(LDFLAGS) main.go

.PHONY: run
run:
	go run main.go

.PHONY: generate
generate: generate-server generate-client

# go-get-tool will 'go get' any package $2 and install it to $1.
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef
