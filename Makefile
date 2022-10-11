VERSION = 0.0.15

IMG_REPO = quay.io/kuberlogic
IMG_LATEST_TAG=latest
OPERATOR_NAME = dynamic-operator
OPERATOR_IMG ?= $(IMG_REPO)/$(OPERATOR_NAME)

APISERVER_NAME = dynamic-apiserver
APISERVER_IMG = $(IMG_REPO)/$(APISERVER_NAME)

CHARGEBEE_INTEGRATION_NAME = chargebee-integration
CHARGEBEE_INTEGRATION_IMG = $(IMG_REPO)/$(CHARGEBEE_INTEGRATION_NAME)

ifeq ($(DEV_BUILD),true)
	VERSION := $(VERSION)-$(shell git rev-list --count $(shell git rev-parse --abbrev-ref HEAD))
endif
COMMIT_SHA = $(shell git rev-parse HEAD)
IMG_SHA_TAG ?= $(VERSION)-$(COMMIT_SHA)

.PHONY: set-version
set-version:
	cd modules/dynamic-operator && \
	OPERATOR_IMG=$(OPERATOR_IMG) \
	APISERVER_IMG=$(APISERVER_IMG) \
	CHARGEBEE_INTEGRATION_IMG=$(CHARGEBEE_INTEGRATION_IMG) \
	VERSION=$(VERSION) \
	$(MAKE) set-version

.PHONY: operator-docker-build
operator-docker-build:
	cd modules/dynamic-operator && \
	$(MAKE) build build-docker-compose-plugin && \
	docker build . \
		--build-arg BIN=bin/manager \
		--build-arg PLUGIN=bin/docker-compose-plugin \
		-t $(OPERATOR_IMG):$(VERSION) \
		-t $(OPERATOR_IMG):$(IMG_SHA_TAG) \
		-t $(OPERATOR_IMG):$(IMG_LATEST_TAG)

.PHONY: apiserver-docker-build
apiserver-docker-build:
	cd modules/dynamic-apiserver && \
	$(MAKE) build && \
	docker build . \
		--build-arg BIN=bin/apiserver \
		-t $(APISERVER_IMG):$(VERSION) \
		-t $(APISERVER_IMG):$(IMG_LATEST_TAG) \
		-t $(APISERVER_IMG):$(IMG_SHA_TAG)


.PHONY: chargebee-docker-build
chargebee-docker-build:
	cd modules/chargebee-integration && \
	$(MAKE) build && \
	docker build . \
		--build-arg BIN=bin/chargebee-integration \
		-t $(CHARGEBEE_INTEGRATION_IMG):$(VERSION) \
		-t $(CHARGEBEE_INTEGRATION_IMG):$(IMG_LATEST_TAG) \
		-t $(CHARGEBEE_INTEGRATION_IMG):$(IMG_SHA_TAG)

.PHONY: docker-build
docker-build: operator-docker-build apiserver-docker-build chargebee-docker-build

.PHONY: docker-push
docker-push:
	docker push $(OPERATOR_IMG):$(VERSION)
	docker push $(OPERATOR_IMG):$(IMG_LATEST_TAG)
	docker push $(APISERVER_IMG):$(VERSION)
	docker push $(APISERVER_IMG):$(IMG_LATEST_TAG)
	docker push $(CHARGEBEE_INTEGRATION_IMG):$(VERSION)
	docker push $(CHARGEBEE_INTEGRATION_IMG):$(IMG_LATEST_TAG)


.PHONY: docker-push-cache
docker-push-cache:
	docker push $(OPERATOR_IMG):$(IMG_SHA_TAG)
	docker push $(APISERVER_IMG):$(IMG_SHA_TAG)
	docker push $(CHARGEBEE_INTEGRATION_IMG):$(IMG_SHA_TAG)

.PHONY: docker-pull-cache
docker-pull-cache:
	docker pull $(OPERATOR_IMG):$(IMG_SHA_TAG)
	docker pull $(APISERVER_IMG):$(IMG_SHA_TAG)
	docker pull $(CHARGEBEE_INTEGRATION_IMG):$(IMG_SHA_TAG)


.PHONY: docker-restore-cache
docker-restore-cache: docker-pull-cache
	docker tag $(OPERATOR_IMG):$(IMG_SHA_TAG) $(OPERATOR_IMG):$(VERSION)
	docker tag $(OPERATOR_IMG):$(IMG_SHA_TAG) $(OPERATOR_IMG):$(IMG_LATEST_TAG)
	docker tag $(APISERVER_IMG):$(IMG_SHA_TAG) $(APISERVER_IMG):$(VERSION)
	docker tag $(APISERVER_IMG):$(IMG_SHA_TAG) $(APISERVER_IMG):$(IMG_LATEST_TAG)
	docker tag $(CHARGEBEE_INTEGRATION_IMG):$(IMG_SHA_TAG) $(CHARGEBEE_INTEGRATION_IMG):$(VERSION)
	docker tag $(CHARGEBEE_INTEGRATION_IMG):$(IMG_SHA_TAG) $(CHARGEBEE_INTEGRATION_IMG):$(IMG_LATEST_TAG)
