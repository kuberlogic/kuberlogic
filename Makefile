.EXPORT_ALL_VARIABLES:

# Current Operator version
VERSION ?= 1.0.1
COMMIT_SHA = $(shell git rev-parse HEAD)

ifeq ($(USE_BUILD),true)
	VERSION := $(VERSION)-$(shell git rev-list --count $(shell git rev-parse --abbrev-ref HEAD))
endif

# private repo for images
IMG_REPO = quay.io/kuberlogic
# default secrets with credentials to private repo (using for mysql/redis)
# for postgresql is using service account
IMG_PULL_SECRET = ""

# Image URL to use all building/pushing image targets
OPERATOR_NAME = operator
OPERATOR_IMG ?= $(IMG_REPO)/$(OPERATOR_NAME):$(VERSION)
OPERATOR_IMG_SHA ?= $(IMG_REPO)/$(OPERATOR_NAME):$(VERSION)-$(COMMIT_SHA)
OPERATOR_IMG_LATEST ?= $(IMG_REPO)/$(OPERATOR_NAME):latest
# updater image name
UPDATER_NAME = updater
UPDATER_IMG ?= $(IMG_REPO)/$(UPDATER_NAME):$(VERSION)
UPDATER_IMG_SHA ?= $(IMG_REPO)/$(UPDATER_NAME):$(VERSION)-$(COMMIT_SHA)
UPDATER_IMG_LATEST ?= $(IMG_REPO)/$(UPDATER_NAME):latest
 # alert receiver image name
ALERT_RECEIVER_NAME = alert-receiver
ALERT_RECEIVER_IMG ?= $(IMG_REPO)/$(ALERT_RECEIVER_NAME):$(VERSION)
ALERT_RECEIVER_IMG_SHA ?= $(IMG_REPO)/$(ALERT_RECEIVER_NAME):$(VERSION)-$(COMMIT_SHA)
ALERT_RECEIVER_IMG_LATEST ?= $(IMG_REPO)/$(ALERT_RECEIVER_NAME):latest
# apiserver image
APISERVER_NAME = apiserver
APISERVER_IMG = $(IMG_REPO)/$(APISERVER_NAME):$(VERSION)
APISERVER_IMG_SHA = $(IMG_REPO)/$(APISERVER_NAME):$(VERSION)-$(COMMIT_SHA)
APISERVER_IMG_LATEST = $(IMG_REPO)/$(APISERVER_NAME):latest

# ui image
UI_NAME = ui
UI_IMG = $(IMG_REPO)/$(UI_NAME):$(VERSION)
UI_IMG_SHA = $(IMG_REPO)/$(UI_NAME):$(VERSION)-$(COMMIT_SHA)
UI_IMG_LATEST = $(IMG_REPO)/$(UI_NAME):latest

# backup image prefix
BACKUP_PREFIX = backup
MYSQL_BACKUP_IMG = $(IMG_REPO)/$(BACKUP_PREFIX)-mysql:$(VERSION)
MYSQL_BACKUP_IMG_SHA = $(IMG_REPO)/$(BACKUP_PREFIX)-mysql:$(VERSION)-$(COMMIT_SHA)
MYSQL_BACKUP_IMG_LATEST = $(IMG_REPO)/$(BACKUP_PREFIX)-mysql:latest
PG_BACKUP_IMG = $(IMG_REPO)/$(BACKUP_PREFIX)-postgresql:$(VERSION)
PG_BACKUP_IMG_SHA = $(IMG_REPO)/$(BACKUP_PREFIX)-postgresql:$(VERSION)-$(COMMIT_SHA)
PG_BACKUP_IMG_LATEST = $(IMG_REPO)/$(BACKUP_PREFIX)-postgresql:latest

# restore from backup image prefix
RESTORE_BACKUP_PREFIX = backup-restore
MYSQL_RESTORE_BACKUP_IMG = $(IMG_REPO)/$(RESTORE_BACKUP_PREFIX)-mysql:$(VERSION)
MYSQL_RESTORE_BACKUP_IMG_SHA = $(IMG_REPO)/$(RESTORE_BACKUP_PREFIX)-mysql:$(VERSION)-$(COMMIT_SHA)
MYSQL_RESTORE_BACKUP_IMG_LATEST = $(IMG_REPO)/$(RESTORE_BACKUP_PREFIX)-mysql:latest
PG_RESTORE_BACKUP_IMG = $(IMG_REPO)/$(RESTORE_BACKUP_PREFIX)-postgresql:$(VERSION)
PG_RESTORE_BACKUP_IMG_SHA = $(IMG_REPO)/$(RESTORE_BACKUP_PREFIX)-postgresql:$(VERSION)-$(COMMIT_SHA)
PG_RESTORE_BACKUP_IMG_LATEST = $(IMG_REPO)/$(RESTORE_BACKUP_PREFIX)-postgresql:latest

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# apiserver section
KUBERLOGIC_AUTH_PROVIDER = none

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif
GOPRIVATE=github.com/kuberlogic

SENTRY_DSN =
NAMESPACE ?= kuberlogic

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
	go run main.go

show-resources:
	kubectl api-resources --verbs=list --namespaced -o name \
      | xargs -n 1 kubectl get --show-kind --ignore-not-found -n $(NAMESPACE)

after-deploy:
	kubectl config set-context --current --namespace=$(NAMESPACE)

# Deploy kuberlogic-operator in the configured Kubernetes cluster in ~/.kube/config
deploy: kustomize manifests deploy-certmanager
	cd config/manager && \
	$(KUSTOMIZE) edit set image operator=$(OPERATOR_IMG) && \
	$(KUSTOMIZE) edit set image controller=$(OPERATOR_IMG) && \
	$(KUSTOMIZE) edit set image apiserver=$(APISERVER_IMG)
	cd config/updater && $(KUSTOMIZE) edit set image updater=$(UPDATER_IMG)
	$(KUSTOMIZE) build config/default | envsubst | kubectl apply -f -
	$(MAKE) after-deploy

undeploy: kustomize undeploy-certmanager
	# need to remove several resources before their operators were removed
	kubectl delete mysqldatabase grafana; \
	kubectl delete mysql grafana; \
	kubectl delete keycloakusers --all-namespaces --all; \
	kubectl delete keycloakclients --all-namespaces --all; \
	kubectl delete keycloakrealms --all-namespaces --all; \
	$(KUSTOMIZE) build config/default | envsubst | kubectl delete -f -

# Install CRDs into a cluster
install: kustomize
	kubectl apply -f config/certmanager/cert-manager-crd.yaml \
	 -f config/keycloak/crd/ \
	 -f config/crd/bases/mysql/presslabs/ \
	 -f config/crd/bases/postgresql/zalando/
	 $(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs into a cluster
uninstall: kustomize
	kubectl delete -f config/certmanager/cert-manager-crd.yaml \
	 -f config/keycloak/crd/ \
	 -f config/crd/bases/mysql/presslabs/ \
	 -f config/crd/bases/postgresql/zalando/
	 $(KUSTOMIZE) build config/crd | kubectl delete -f -

deploy-certmanager:
	kubectl apply -f config/certmanager/cert-manager.yaml
	kubectl wait -n cert-manager --for=condition=Ready pods --all --timeout=5m
	sleep 10 # need to wait cert-manager deployments

undeploy-certmanager:
	kubectl delete -f config/certmanager/cert-manager.yaml

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	cd modules/operator; \
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=../../charts/crds/templates ;\

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
	docker build modules/operator \
		-t $(OPERATOR_IMG) \
		-t $(OPERATOR_IMG_SHA) \
		-t $(OPERATOR_IMG_LATEST) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(shell date +"%d-%m-%yT%T%z") \
		--build-arg REVISION=$(COMMIT_SHA)

installer-build:
	@cd modules/installer; \
	VERSION=$(VERSION) \
	BUILD_TIME=$(shell date +"%d-%m-%yT%T%z") \
	REVISION=$(COMMIT_SHA) \
	$(MAKE) build

updater-build:
	docker build . \
	-f updater.Dockerfile \
	-t $(UPDATER_IMG) \
	-t $(UPDATER_IMG_SHA) \
	-t $(UPDATER_IMG_LATEST)

alert-receiver-build:
	docker build . \
	-f alert-receiver.Dockerfile \
	-t $(ALERT_RECEIVER_IMG) \
	-t $(ALERT_RECEIVER_IMG_SHA) \
	-t $(ALERT_RECEIVER_IMG_LATEST) \

apiserver-build:
	echo "Building apiserver image"
	docker build . -f apiserver.Dockerfile \
	-t $(APISERVER_IMG) \
	-t $(APISERVER_IMG_SHA) \
	-t $(APISERVER_IMG_LATEST) \
	--build-arg VERSION=$(VERSION) \
	--build-arg BUILD_TIME=$(shell date +"%d-%m-%yT%T%z") \
	--build-arg REVISION=$(COMMIT_SHA)

build-tests: gen test
	echo "Building tests image"
	docker build . -f Dockerfile.tests -t $(TESTS_IMG) -t $(TESTS_IMG_LATEST) .

push-tests:
	docker push $(TESTS_IMG)
	docker push $(TESTS_IMG_LATEST)

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
	docker push $(APISERVER_IMG)
	docker push $(APISERVER_IMG_LATEST)

tests-build: apiserver-gen
	echo "Building tests image"
	docker build . -f tests.Dockerfile -t $(TESTS_IMG) -t $(TESTS_IMG_LATEST)

tests-push:
	docker push $(TESTS_IMG)
	docker push $(TESTS_IMG_LATEST)

mark-executable:
	chmod +x $(shell find backup/ -iname *.sh | xargs)

backup-build: mark-executable
	docker build backup/mysql/ \
	-t $(MYSQL_BACKUP_IMG) \
	-t $(MYSQL_BACKUP_IMG_SHA) \
	-t $(MYSQL_BACKUP_IMG_LATEST)
	docker build backup/postgres/ \
	-t $(PG_BACKUP_IMG) \
	-t $(PG_BACKUP_IMG_SHA) \
	-t $(PG_BACKUP_IMG_LATEST)

backup-push:
	docker push $(MYSQL_BACKUP_IMG)
	docker push $(MYSQL_BACKUP_IMG_LATEST)
	docker push $(PG_BACKUP_IMG)
	docker push $(PG_BACKUP_IMG_LATEST)

restore-build: mark-executable
	docker build backup/restore/mysql/ \
	-t $(MYSQL_RESTORE_BACKUP_IMG) \
	-t $(MYSQL_RESTORE_BACKUP_IMG_SHA) \
	-t $(MYSQL_RESTORE_BACKUP_IMG_LATEST)
	docker build backup/restore/postgres/ \
	-t $(PG_RESTORE_BACKUP_IMG) \
	-t $(PG_RESTORE_BACKUP_IMG_SHA) \
	-t $(PG_RESTORE_BACKUP_IMG_LATEST)

restore-push:
	docker push $(MYSQL_RESTORE_BACKUP_IMG)
	docker push $(MYSQL_RESTORE_BACKUP_IMG_LATEST)
	docker push $(PG_RESTORE_BACKUP_IMG)
	docker push $(PG_RESTORE_BACKUP_IMG_LATEST)

ui-build:
	docker build modules/ui \
	-t $(UI_IMG) \
	-t $(UI_IMG_SHA) \
	-t $(UI_IMG_LATEST) \

ui-push:
	docker push $(UI_IMG)
	docker push $(UI_IMG_LATEST)

docker-push: operator-push apiserver-push updater-push alert-receiver-push backup-push restore-push ui-push
docker-build: operator-build apiserver-build updater-build alert-receiver-build backup-build restore-build ui-build

docker-push-cache:
	for image in \
		$(OPERATOR_IMG_SHA) \
		$(APISERVER_IMG_SHA) \
		$(UPDATER_IMG_SHA) \
		$(ALERT_RECEIVER_IMG_SHA) \
		$(UI_IMG_SHA) \
		$(MYSQL_BACKUP_IMG_SHA) \
		$(PG_BACKUP_IMG_SHA) \
		$(MYSQL_RESTORE_BACKUP_IMG_SHA) \
		$(PG_RESTORE_BACKUP_IMG_SHA) \
		; do \
			docker push $${image}; \
	done

docker-pull-cache:
	for image in \
		$(OPERATOR_IMG_SHA) \
		$(APISERVER_IMG_SHA) \
		$(UPDATER_IMG_SHA) \
		$(ALERT_RECEIVER_IMG_SHA) \
		$(UI_IMG_SHA) \
		$(MYSQL_BACKUP_IMG_SHA) \
		$(PG_BACKUP_IMG_SHA) \
		$(MYSQL_RESTORE_BACKUP_IMG_SHA) \
		$(PG_RESTORE_BACKUP_IMG_SHA) \
		; do \
			docker pull $${image}; \
	done

docker-restore-cache: docker-pull-cache
	docker tag $(OPERATOR_IMG_SHA) $(OPERATOR_IMG)
	docker tag $(OPERATOR_IMG_SHA) $(OPERATOR_IMG_LATEST)
	docker tag $(APISERVER_IMG_SHA) $(APISERVER_IMG)
	docker tag $(APISERVER_IMG_SHA) $(APISERVER_IMG_LATEST)
	docker tag $(UPDATER_IMG_SHA) $(UPDATER_IMG)
	docker tag $(UPDATER_IMG_SHA) $(UPDATER_IMG_LATEST)
	docker tag $(ALERT_RECEIVER_IMG_SHA) $(ALERT_RECEIVER_IMG)
	docker tag $(ALERT_RECEIVER_IMG_SHA) $(ALERT_RECEIVER_IMG_LATEST)
	docker tag $(UI_IMG_SHA) $(UI_IMG)
	docker tag $(UI_IMG_SHA) $(UI_IMG_LATEST)
	docker tag $(MYSQL_BACKUP_IMG_SHA) $(MYSQL_BACKUP_IMG)
	docker tag $(MYSQL_BACKUP_IMG_SHA) $(MYSQL_BACKUP_IMG_LATEST)
	docker tag $(PG_BACKUP_IMG_SHA) $(PG_BACKUP_IMG)
	docker tag $(PG_BACKUP_IMG_SHA) $(PG_BACKUP_IMG_LATEST)
	docker tag $(MYSQL_RESTORE_BACKUP_IMG_SHA) $(MYSQL_RESTORE_BACKUP_IMG)
	docker tag $(MYSQL_RESTORE_BACKUP_IMG_SHA) $(MYSQL_RESTORE_BACKUP_IMG_LATEST)
	docker tag $(PG_RESTORE_BACKUP_IMG_SHA) $(PG_RESTORE_BACKUP_IMG)
	docker tag $(PG_RESTORE_BACKUP_IMG_SHA) $(PG_RESTORE_BACKUP_IMG_LATEST)

refresh-go-sum:
	for module in operator updater alert-receiver apiserver installer; do \
  		cd ./modules/$${module}; \
  		go clean -modcache; \
  		go mod tidy; \
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
