.EXPORT_ALL_VARIABLES:

KUBERLOGIC_AUTH_PROVIDER = none

# private repo for images
IMG_REPO = quay.io/kuberlogic

# default secrets with credetials to private repo (using for mysql/redis)
# for postgresql is using service account
IMG_PULL_SECRET = kuberlogic-registry

KUBERLOGIC_APISERVER_LOG=/var/log/kuberlogic.apiserver.log
KUBERLOGIC_OPERATOR_LOG=/var/log/kuberlogic.operator.log

DEBUG_LOGS = True

ARGS ?= "-short"

INTERNAL_MINIO_IP=$(shell kubectl get pod -l app=minio -o go-template="{{range .items}}{{.status.podIP}}{{end}}")

test:
	go test -parallel 2 -v -coverprofile=coverage.out -coverpkg="github.com/kuberlogic/..." -run "$(TEST)" -timeout 60m $(ARGS) ./...
	mkdir -p html/
	for pkg in apiserver operator; do \
		echo "mode: set" > $${pkg}.out; \
		cat coverage.out | grep "github.com/kuberlogic/$${pkg}" >> $${pkg}.out; \
		go tool cover -html=$${pkg}.out -o=html/$${pkg}.html; \
	done

build-minio:
	docker build minio/ -t kuberlogic/minio

run-minio: build-minio
	kubectl apply -f minio/deployment.yaml

internal-minio-ip:
	@echo $$INTERNAL_MINIO_IP

destroy-minio:
	kubectl delete -f minio/deployment