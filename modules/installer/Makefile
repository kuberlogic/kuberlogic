CHARTS_DIR="../../charts"

helm-deps:
	cd $(CHARTS_DIR) && \
	helm dependency update monitoring

helm-lint:
	cd $(CHARTS_DIR) && \
	helm lint crds && \
	helm lint nginx-ingress-controller && \
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
	helm package nginx-ingress-controller && \
	helm package cert-manager && \
	helm package kuberlogic-ui && \
	helm package kuberlogic-apiserver && \
	helm package keycloak-operator && \
	helm package kuberlogic-keycloak && \
	helm package monitoring && \
	helm package mysql-operator && \
	helm package postgres-operator && \
	helm package kuberlogic-operator

helm-embed: helm-package
	rm kli/helm-installer/*tgz ; \
	mv $(CHARTS_DIR)/*.tgz kli/helm-installer

test:
	go vet

compile: test
	go build \
	    -ldflags " \
        -X github.com/kuberlogic/operator/modules/installer/cmd.sha1ver=$(REVISION) \
        -X github.com/kuberlogic/operator/modules/installer/cmd.buildTime=$(BUILD_TIME) \
        -X github.com/kuberlogic/operator/modules/installer/cmd.version=$(VERSION)  \
        -X github.com/kuberlogic/operator/modules/installer/kli/helm-installer.apiserverTag=$(VERSION)  \
        -X github.com/kuberlogic/operator/modules/installer/kli/helm-installer.operatorTag=$(VERSION)"  \
        -o kuberlogic-installer

build: helm-embed compile
