.EXPORT_ALL_VARIABLES:

# Current Operator version
VERSION ?= 0.0.23
# Default bundle image tag
BUNDLE_IMG ?= kuberlogic-operator:$(VERSION)
# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# private repo for images
IMG_REPO = quay.io/kuberlogic
# default secrets with credentials to private repo (using for mysql/redis)
# for postgresql is using service account
IMG_PULL_SECRET = kuberlogic-registry

# Image URL to use all building/pushing image targets
OPERATOR_NAME = operator
IMG ?= $(IMG_REPO)/$(OPERATOR_NAME):$(VERSION)
# updater image name
UPDATER_NAME = updater
UPDATER_IMG ?= $(IMG_REPO)/$(UPDATER_NAME):$(VERSION)
 # alert receiver image name
ALERT_RECEIVER_NAME = alert-receiver
ALERT_RECEIVER_IMG ?= $(IMG_REPO)/$(ALERT_RECEIVER_NAME):$(VERSION)
# backup image prefix
BACKUP_PREFIX = backup
# restore from backup image prefix
RESTORE_PREFIX = backup-restore

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif
GOPRIVATE=github.com/kuberlogic

SENTRY_DSN =

all: manager

# Run tests
test: generate fmt vet manifests
	cd modules/operator; \
	go test ./... -coverprofile cover.out ;\

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	cd modules/operator ;\
	go run main.go ;\

# Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

# Deploy kuberlogic-operator in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image operator=$(IMG)
	cd config/updater && $(KUSTOMIZE) edit set image updater=$(UPDATER_IMG)
	$(KUSTOMIZE) build config/default | kubectl apply -f -

undeploy:
	$(KUSTOMIZE) build config/default | kubectl delete -f -

deploy-dependencies: manifests kustomize
	$(KUSTOMIZE) build config/dependencies | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	cd modules/operator; \
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=../../config/crd/bases output:webhook:artifacts:config=../../config/webhook ;\

# Run go fmt against code
fmt:
	cd modules/operator ;\
	go fmt ./... ;\

# Run go vet against code
vet:
	cd modules/operator ; \
	go vet ./... ; \


# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="modules/operator/..."

# Build the operator images
operator-build:
	echo "Building images"
	docker build modules/operator -t $(IMG)
	docker build modules/updater -t $(UPDATER_IMG)
	docker build modules/alert-receiver -t $(ALERT_RECEIVER_IMG)

# Push operator images
operator-push:
	docker push $(IMG)
	docker push $(UPDATER_IMG)
	docker push $(ALERT_RECEIVER_IMG)

mark-executable:
	chmod +x $(shell find backup/ -iname *.sh | xargs)

# Build backup images
backup-build: mark-executable
	docker build backup/mysql/ -t $(IMG_REPO)/$(BACKUP_PREFIX)-mysql:$(VERSION) -t $(IMG_REPO)/$(BACKUP_PREFIX)-mysql:latest
	docker build backup/postgres/ -t $(IMG_REPO)/$(BACKUP_PREFIX)-postgresql:$(VERSION) -t $(IMG_REPO)/$(BACKUP_PREFIX)-postgresql:latest

# Push backup images
backup-push:
	docker push $(IMG_REPO)/$(BACKUP_PREFIX)-mysql:$(VERSION)
	docker push $(IMG_REPO)/$(BACKUP_PREFIX)-mysql:latest
	docker push $(IMG_REPO)/$(BACKUP_PREFIX)-postgresql:$(VERSION)
	docker push $(IMG_REPO)/$(BACKUP_PREFIX)-postgresql:latest

# Build backup restore images
restore-build: mark-executable
	docker build backup/restore/mysql/ -t $(IMG_REPO)/$(RESTORE_PREFIX)-mysql:$(VERSION) -t $(IMG_REPO)/$(RESTORE_PREFIX)-mysql:latest
	docker build backup/restore/postgres/ -t $(IMG_REPO)/$(RESTORE_PREFIX)-postgresql:$(VERSION) -t $(IMG_REPO)/$(RESTORE_PREFIX)-postgresql:latest

# Push backup restore images
restore-push:
	docker push $(IMG_REPO)/$(RESTORE_PREFIX)-mysql:$(VERSION)
	docker push $(IMG_REPO)/$(RESTORE_PREFIX)-mysql:latest
	docker push $(IMG_REPO)/$(RESTORE_PREFIX)-postgresql:$(VERSION)
	docker push $(IMG_REPO)/$(RESTORE_PREFIX)-postgresql:latest

docker-build: operator-build backup-build restore-build
	#

docker-push: operator-push backup-push restore-push
	#

refresh-go-sum:
	for module in operator updater alert-receiver watcher apiserver; do \
  		cd ./modules/$${module}; \
  		go clean -modcache; \
  		go mod tidy; \
  		cd -; \
	done

bump-operator-version:
	set -o errexit; \
	for module in updater alert-receiver watcher apiserver; do \
  		echo "Entering into" $${module}; \
	  	cd ./modules/$${module}; \
	  	go mod edit -droprequire github.com/kuberlogic/operator/modules/operator go.mod; \
  		go get github.com/kuberlogic/operator/modules/operator@${BRANCH}; \
  		cd -; \
	done

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.5.0)

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v4@v4.1.2)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef