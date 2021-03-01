.EXPORT_ALL_VARIABLES:

# Current Operator version
VERSION ?= 0.0.18
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
IMG_PULL_SECRET = gitlab-registry

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

all: manager

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

# Deploy kuberlogic-operator in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image operator=${IMG}
	cd config/updater && $(KUSTOMIZE) edit set image updater=$(UPDATER_IMG)
	$(KUSTOMIZE) build config/default | kubectl apply -f -

undeploy:
	$(KUSTOMIZE) build config/default | kubectl delete -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the operator images
operator-build:
	echo "Building images"
	docker build . -t ${IMG}
	docker build updater/ -t ${UPDATER_IMG}
	docker build alert-receiver/ -t ${ALERT_RECEIVER_IMG}

# Push operator images
operator-push:
	docker push ${IMG}
	docker push ${UPDATER_IMG}
	docker push ${ALERT_RECEIVER_IMG}

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

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

kustomize:
ifeq (, $(shell which kustomize))
	@{ \
	set -e ;\
	KUSTOMIZE_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$KUSTOMIZE_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/kustomize/kustomize/v3@v3.5.4 ;\
	rm -rf $$KUSTOMIZE_GEN_TMP_DIR ;\
	}
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif

# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: manifests
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image operator=$(IMG)
	cd config/updater && $(KUSTOMIZE) edit set image updater=$(UPDATER_IMG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle

# Build the bundle image.
.PHONY: bundle-build
bundle-build:
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .
