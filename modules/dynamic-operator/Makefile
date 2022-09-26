# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.22


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
	NAMESPACE="kuberlogic" \
	SERVICE_ACCOUNT="default" \
	BACKUPS_ENABLED="true" \
	KL_PLUGIN_DOCKER_COMPOSE_FILE="$(PROJECT_DIR)/config/manager/docker-compose.yaml" \
	go test -timeout 30m -v ./... -coverprofile cover.out ${ARGS}

cluster-test: envtest deploy-velero
	USE_EXISTING_CLUSTER=true \
	ACK_GINKGO_DEPRECATIONS=1.16.5 \
	NAMESPACE=kuberlogic \
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" \
	go test -v ./... -coverprofile cover.out ${ARGS}

##@ Build

build-docker-compose-plugin: export LDFLAGS="-X main.ver=$(VERSION)"
build-docker-compose-plugin:
	cd plugins/docker-compose/ && \
	GOARCH=amd64 GOOS=linux go build -a -o ../../bin/docker-compose-plugin -ldflags=$(LDFLAGS) .

.PHONY: build
build: export CGO_ENABLED=0
build: export GO111MODULE=on
build: export LDFLAGS="-X main.ver=$(VERSION)"
build: generate fmt vet test build-docker-compose-plugin
	GOARCH=amd64 GOOS=linux go build -a -o bin/manager -ldflags=$(LDFLAGS) main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	NAMESPACE=kuberlogic SENTRY_DSN=${SENTRY_DSN} \
	go run ./main.go ${ARGS}

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: set-version
set-version: kustomize
	cd config/manager && \
		$(KUSTOMIZE) edit set image controller=$(OPERATOR_IMG):$(VERSION) && \
		$(KUSTOMIZE) edit set image apiserver=$(APISERVER_IMG):$(VERSION) && \
		$(KUSTOMIZE) edit set image chargebee-integration=$(CHARGEBEE_INTEGRATION_IMG):$(VERSION)

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy-certmanager
deploy-certmanager: kustomize
	$(KUSTOMIZE) build config/cert-manager | kubectl apply -f -

.PHONY: deploy
deploy: set-version
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy-certmanager
undeploy-certmanager:
	$(KUSTOMIZE) build config/cert-manager | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: undeploy
undeploy: kustomize ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

deploy-velero:
	kubectl apply -f config/velero/crd
	kubectl apply -k config/velero
	kubectl -n velero wait --timeout=120s --for=condition=Ready pod -l app.kubernetes.io/instance=velero

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
	kubectl get secret --namespace=kuberlogic webhook-server-cert -o=jsonpath="{.data['tls\.key']}" | base64 -d >  /tmp/k8s-webhook-server/serving-certs/tls.key
	kubectl get secret --namespace=kuberlogic webhook-server-cert -o=jsonpath="{.data['tls\.crt']}" | base64 -d >  /tmp/k8s-webhook-server/serving-certs/tls.crt
