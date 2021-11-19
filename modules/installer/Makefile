CHARTS_DIR="../../charts"
GOARCH=amd64
GOOS=linux
HELM=helm --debug

helm-deps:
	cd $(CHARTS_DIR) && \
	$(HELM) dependency update monitoring

helm-lint:
	cd $(CHARTS_DIR) && \
	$(HELM) lint crds && \
	$(HELM) lint kong && \
	$(HELM) lint kuberlogic-ingress-controller && \
	$(HELM) lint kuberlogic-ui && \
	$(HELM) lint kuberlogic-apiserver && \
	$(HELM) lint keycloak-operator && \
	$(HELM) lint kuberlogic-keycloak && \
	$(HELM) lint monitoring && \
	$(HELM) lint mysql-operator && \
	$(HELM) lint postgres-operator && \
	$(HELM) lint kuberlogic-operator

helm-package: helm-deps helm-lint
	cd $(CHARTS_DIR) && find . -name "*.tgz" -delete ; \
	$(HELM) package crds && \
	$(HELM) package kong && \
	$(HELM) package kuberlogic-ingress-controller && \
	$(HELM) package cert-manager && \
	$(HELM) package keycloak-operator && \
	$(HELM) package kuberlogic-keycloak && \
	$(HELM) package monitoring && \
	$(HELM) package mysql-operator && \
	$(HELM) package postgres-operator && \
	$(HELM) package kuberlogic-ui --app-version $(VERSION) && \
    $(HELM) package kuberlogic-apiserver --app-version $(VERSION) && \
	$(HELM) package kuberlogic-operator --app-version $(VERSION)

helm-embed: helm-package
	rm kli/$(HELM)-installer/*tgz ; \
	mv $(CHARTS_DIR)/*.tgz kli/helm-installer

test: helm-embed
	go vet && \
	go test ./... -race -covermode=atomic -coverprofile=unit-coverage.out

compile: helm-embed
	mkdir -p bin && \
	go mod vendor && \
	go build \
		-mod=vendor \
	    -ldflags " \
        -X github.com/kuberlogic/kuberlogic/modules/installer/cmd.sha1ver=$(REVISION) \
        -X github.com/kuberlogic/kuberlogic/modules/installer/cmd.buildTime=$(BUILD_TIME) \
        -X github.com/kuberlogic/kuberlogic/modules/installer/cmd.version=$(VERSION)"  \
        -o bin/kuberlogic-installer-$(GOOS)-$(GOARCH)

release:
	$(MAKE) compile GOOS=darwin
	$(MAKE) compile GOOS=linux
