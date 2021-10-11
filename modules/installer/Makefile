CHARTS_DIR="../../charts"
GOARCH=amd64
GOOS=linux

helm-deps:
	cd $(CHARTS_DIR) && \
	helm dependency update monitoring

helm-lint:
	cd $(CHARTS_DIR) && \
	helm lint crds && \
	helm lint kong && \
	helm lint kuberlogic-ingress-controller && \
	helm lint kuberlogic-ui && \
	helm lint kuberlogic-apiserver && \
	helm lint keycloak-operator && \
	helm lint kuberlogic-keycloak && \
	helm lint monitoring && \
	helm lint mysql-operator && \
	helm lint postgres-operator && \
	helm lint kuberlogic-operator

helm-package: helm-deps helm-lint
	cd $(CHARTS_DIR) && find . -name "*.tgz" -delete ; \
	helm package crds && \
	helm package kong && \
	helm package kuberlogic-ingress-controller && \
	helm package cert-manager && \
	helm package keycloak-operator && \
	helm package kuberlogic-keycloak && \
	helm package monitoring && \
	helm package mysql-operator && \
	helm package postgres-operator && \
	helm package kuberlogic-ui --app-version $(VERSION) && \
    helm package kuberlogic-apiserver --app-version $(VERSION) && \
	helm package kuberlogic-operator --app-version $(VERSION)

helm-embed: helm-package
	rm kli/helm-installer/*tgz ; \
	mv $(CHARTS_DIR)/*.tgz kli/helm-installer

test: helm-embed
	go vet

compile: test
	mkdir -p bin && \
	go build \
	    -ldflags " \
        -X github.com/kuberlogic/kuberlogic/modules/installer/cmd.sha1ver=$(REVISION) \
        -X github.com/kuberlogic/kuberlogic/modules/installer/cmd.buildTime=$(BUILD_TIME) \
        -X github.com/kuberlogic/kuberlogic/modules/installer/cmd.version=$(VERSION)"  \
        -o bin/kuberlogic-installer-$(GOOS)-$(GOARCH)

release:
	$(MAKE) compile GOOS=darwin
	$(MAKE) compile GOOS=linux
