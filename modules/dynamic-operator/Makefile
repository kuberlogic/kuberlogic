VERSION = 0.0.11

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.22

IMG_REPO = quay.io/kuberlogic
IMG_LATEST_TAG=latest
OPERATOR_NAME = dynamic-operator
OPERATOR_IMG ?= $(IMG_REPO)/$(OPERATOR_NAME)

APISERVER_NAME = dynamic-apiserver
APISERVER_IMG = $(IMG_REPO)/$(APISERVER_NAME)

CHARGEBEE_INTEGRATION_NAME = chargebee-integration
CHARGEBEE_INTEGRATION_IMG = $(IMG_REPO)/$(CHARGEBEE_INTEGRATION_NAME)

PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))

SENTRY_DSN =

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

ifeq ($(USE_BUILD),true)
	VERSION := $(VERSION)-$(shell git rev-list --count $(shell git rev-parse --abbrev-ref HEAD))
endif
COMMIT_SHA = $(shell git rev-parse HEAD)
IMG_SHA_TAG ?= $(VERSION)-$(COMMIT_SHA)

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

build-docker-compose-plugin:
	cd plugins/docker-compose/ && go build -a -o ../../bin/docker-compose-plugin .

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet build-docker-compose-plugin envtest ## Run tests.
	ACK_GINKGO_DEPRECATIONS=1.16.5 \
	PLUGINS="{docker-compose,$(PROJECT_DIR)/bin/docker-compose-plugin}" \
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" \
	NAMESPACE="default" \
	SERVICE_ACCOUNT="default" \
	BACKUPS_ENABLED="true" \
	KL_PLUGIN_DOCKER_COMPOSE_FILE="$(PROJECT_DIR)/config/manager/docker-compose.yaml" \
	go test -v ./... -coverprofile cover.out ${ARGS}

##@ Build

.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	NAMESPACE=kuberlogic \
	SENTRY_DSN=${SENTRY_DSN} \
	go run ./main.go ${ARGS}

.PHONY: docker-build
docker-build: export CGO_ENABLED=0
docker-build: export GOOS=linux
docker-build: export GOARCH=amd64
docker-build: export GO111MODULE=on
docker-build: test ## Build docker image with the manager.
	go build -a -o bin/manager main.go
	cd plugins/docker-compose && go build -a -o ../../bin/docker-compose-plugin . && cd -
	docker build . \
		--build-arg BIN=bin/manager \
		--build-arg PLUGIN=bin/docker-compose-plugin \
		-t $(OPERATOR_IMG):$(VERSION) \
		-t $(OPERATOR_IMG):$(IMG_SHA_TAG) \
		-t $(OPERATOR_IMG):$(IMG_LATEST_TAG)

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push $(OPERATOR_IMG):$(VERSION)
	docker push $(OPERATOR_IMG):$(IMG_LATEST_TAG)

.PHONY: docker-push-cache
docker-push-cache:
	docker push $(OPERATOR_IMG):$(IMG_SHA_TAG)

.PHONY: docker-pull-cache
docker-pull-cache:
	docker pull $(OPERATOR_IMG):$(IMG_SHA_TAG)

.PHONY: docker-restore-cache
docker-restore-cache: docker-pull-cache
	docker tag $(OPERATOR_IMG):$(IMG_SHA_TAG) $(OPERATOR_IMG):$(VERSION)
	docker tag $(OPERATOR_IMG):$(IMG_SHA_TAG) $(OPERATOR_IMG):$(IMG_LATEST_TAG)

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy-certmanager
deploy-certmanager: kustomize
	$(KUSTOMIZE) build config/cert-manager | kubectl apply -f -

.PHONY: set-version
set-version: kustomize
	cd config/manager && \
		$(KUSTOMIZE) edit set image controller=$(OPERATOR_IMG):$(VERSION) && \
		$(KUSTOMIZE) edit set image apiserver=$(APISERVER_IMG):$(VERSION) && \
		$(KUSTOMIZE) edit set image chargebee-integration=$(CHARGEBEE_INTEGRATION_IMG):$(VERSION)

.PHONY: deploy
deploy: set-version
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy-certmanager
undeploy-certmanager:
	$(KUSTOMIZE) build config/cert-manager | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: undeploy
undeploy: kustomize ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
.PHONY: controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.7.0)

KUSTOMIZE = $(shell pwd)/bin/kustomize
.PHONY: kustomize
kustomize: ## Download kustomize locally if necessary.
	@echo $(shell bash scripts/download-kustomize.sh v4.5.4 $(KUSTOMIZE))

ENVTEST = $(shell pwd)/bin/setup-envtest
.PHONY: envtest
envtest: ## Download envtest-setup locally if necessary.
	$(call go-get-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)

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

.PHONY: patch-webhook-endpoint
patch-webhook-endpoint:
	MINIKUBE_IP=$$(minikube ip) envsubst < config/webhook/patched-endpoint.yaml | kubectl apply -f -

.PHONY: generate-local-webhook-certs
generate-local-webhook-certs:
	mkdir -p /tmp/k8s-webhook-server/serving-certs/; \
	kubectl get secret webhook-server-cert -o=jsonpath="{.data['tls\.key']}" | base64 -d >  /tmp/k8s-webhook-server/serving-certs/tls.key
	kubectl get secret webhook-server-cert -o=jsonpath="{.data['tls\.crt']}" | base64 -d >  /tmp/k8s-webhook-server/serving-certs/tls.crt
