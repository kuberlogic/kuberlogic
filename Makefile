.EXPORT_ALL_VARIABLES:

# Current Operator version
VERSION ?= 0.0.28

# private repo for images
IMG_REPO = quay.io/kuberlogic
# default secrets with credentials to private repo (using for mysql/redis)
# for postgresql is using service account
IMG_PULL_SECRET = kuberlogic-registry

# Image URL to use all building/pushing image targets
OPERATOR_NAME = operator
OPERATOR_IMG ?= $(IMG_REPO)/$(OPERATOR_NAME):$(VERSION)
OPERATOR_IMG_LATEST ?= $(IMG_REPO)/$(OPERATOR_NAME):latest
# updater image name
UPDATER_NAME = updater
UPDATER_IMG ?= $(IMG_REPO)/$(UPDATER_NAME):$(VERSION)
UPDATER_IMG_LATEST ?= $(IMG_REPO)/$(UPDATER_NAME):latest
 # alert receiver image name
ALERT_RECEIVER_NAME = alert-receiver
ALERT_RECEIVER_IMG ?= $(IMG_REPO)/$(ALERT_RECEIVER_NAME):$(VERSION)
ALERT_RECEIVER_IMG_LATEST ?= $(IMG_REPO)/$(ALERT_RECEIVER_NAME):latest
# tests
IMG_TESTS ?= $(IMG_REPO)/integration-tests:$(VERSION)
IMG_TESTS_LATEST ?= $(IMG_REPO)/integration-tests:latest

# backup image prefix
BACKUP_PREFIX = backup
# restore from backup image prefix
RESTORE_PREFIX = backup-restore

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# apiserver section
KUBERLOGIC_AUTH_PROVIDER = none

APISERVER_NAME = apiserver
IMG_APISERVER = $(IMG_REPO)/$(APISERVER_NAME):$(VERSION)
IMG_APISERVER_LATEST = $(IMG_REPO)/$(APISERVER_NAME):latest
IMG_TESTS ?= $(IMG_REPO)/integration-tests:$(VERSION)
IMG_TESTS_LATEST ?= $(IMG_REPO)/integration-tests:latest

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
	cd config/manager && $(KUSTOMIZE) edit set image operator=$(OPERATOR_IMG)
	cd config/updater && $(KUSTOMIZE) edit set image updater=$(UPDATER_IMG)
	$(KUSTOMIZE) build config/default | kubectl apply -f -

undeploy: kustomize
	$(KUSTOMIZE) build config/default | kubectl delete -f -

deploy-requirements: kustomize
	# cert-manager included non-crd resources
	# it can not configured with kustomization due to tight integrate with namespace=cert-manager
	# (see cert-manager.yaml for the details)
	for module in config/certmanager/cert-manager.yaml \
				  config/keycloak/keycloak*.yaml; do \
  		kubectl apply -f $${module} ;\
  	done; \
  	# waiting for cert-manager-webhook endpoint assign an ip address \
  	kubectl wait -n cert-manager --for=condition=Ready pods --all --timeout=5m

undeploy-requirements: kustomize
	for module in config/certmanager/cert-manager.yaml \
				  config/keycloak/crd.yaml; do \
  		kubectl delete -f $${module} ;\
  	done

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	cd modules/operator; \
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=../../config/crd/bases output:webhook:artifacts:config=../../config/webhook ;\

# Run go fmt against code
fmt:
	for module in operator apiserver; do \
		cd ./modules/$${module}; \
		go fmt ./... ;\
	done

# Run go vet against code
vet:
	for module in operator apiserver; do \
		cd ./modules/$${module}; \
		go vet ./... ; \
	done


# Generate code
generate: controller-gen
	cd modules/operator ;\
	$(CONTROLLER_GEN) object paths="./..." output:dir="./api/v1"

# Build the  images
operator-build:
	echo "Building images"
	docker build modules/operator -t $(OPERATOR_IMG) -t $(OPERATOR_IMG_LATEST)

updater-build:
	docker build -f updater.Dockerfile -t $(UPDATER_IMG) -t $(UPDATER_IMG_LATEST) .

alert-receiver-build:
	docker build -f alert-receiver.Dockerfile -t $(ALERT_RECEIVER_IMG) -t $(ALERT_RECEIVER_IMG_LATEST) .

apiserver-build:
	echo "Building apiserver image"
	docker build -f apiserver.Dockerfile -t $(IMG_APISERVER) -t $(IMG_APISERVER_LATEST) .

build-tests: gen test
	echo "Building tests image"
	docker build . -f Dockerfile.tests -t $(IMG_TESTS) -t $(IMG_TESTS_LATEST) .

push-tests:
	docker push $(IMG_TESTS)
	docker push $(IMG_TESTS_LATEST)

# Push images
operator-push:
	docker push $(OPERATOR_IMG)
	docker push $(OPERATOR_IMG_LATEST)

updater-push:
	docker push $(UPDATER_IMG)
	docker push $(UPDATER_IMG_LATEST)

alert-receiver-push:
	docker push $(ALERT_RECEIVER_IMG)
	docker push $(ALERT_RECEIVER_IMG_LATEST)

apiserver-push:
	docker push $(IMG_APISERVER)
	docker push $(IMG_APISERVER_LATEST)

tests-build: apiserver-gen
	echo "Building tests image"
	docker build . -f tests.Dockerfile -t $(IMG_TESTS) -t $(IMG_TESTS_LATEST)

tests-push:
	docker push $(IMG_TESTS)
	docker push $(IMG_TESTS_LATEST)

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

docker-build: operator-build apiserver-build updater-build alert-receiver-build backup-build restore-build
	#

docker-push: operator-push apiserver-push updater-push alert-receiver-push backup-push restore-push
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
	for module in updater alert-receiver apiserver; do \
  		echo "Entering into" $${module}; \
	  	cd ./modules/$${module}; \
	  	go mod edit -droprequire github.com/kuberlogic/operator/modules/operator go.mod; \
  		go get github.com/kuberlogic/operator/modules/operator@${BRANCH}; \
  		cd -; \
	done

apiserver-clean:
	@cd modules/apiserver
	@rm -rf ./cmd internal/generated
	@mkdir -p cmd
	@mkdir -p internal/generated
	@mkdir -p internal/app
	@mkdir -p internal/config

apiserver-gen: apiserver-clean
	cd modules/apiserver; \
	swagger generate server \
		-f openapi.yaml \
		-t internal/generated \
		-C swagger-templates/default-server.yml \
		-P models.Principal \
		--template-dir swagger-templates/templates \
		--name kuberlogic

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
