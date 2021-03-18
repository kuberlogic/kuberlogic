.EXPORT_ALL_VARIABLES:

KUBERLOGIC_AUTH_PROVIDER = none

# private repo for images
IMG_REPO = quay.io/kuberlogic
IMG = $(IMG_REPO)/minio:1.0.0

# default secrets with credetials to private repo (using for mysql/redis)
# for postgresql is using service account
IMG_PULL_SECRET = kuberlogic-registry

KUBERLOGIC_APISERVER_LOG = kuberlogic.apiserver.log
KUBERLOGIC_OPERATOR_LOG = kuberlogic.operator.log

KUBERLOGIC_DEBUGLOGS = True

ARGS ?= -short

INTERNAL_MINIO_IP=$(shell kubectl get pod -l app=minio -o go-template="{{range .items}}{{.status.podIP}}{{end}}")
SERVICE_TYPE = $(shell bash -c "./service-type.sh $(RUN)")

# access/secret also added in the deployment.yaml
MINIO_ACCESS_KEY = BKIKJAA5BMMU2RHO6IBB
MINIO_SECRET_KEY = V7f1CwQqAcwo80UEIJEjc5gVQUSSx5ohQ9GSrr12
MINIO_ENDPOINT = http://$(INTERNAL_MINIO_IP):9000
BUCKET ?= test

OPERATOR_SENTRY_DSN = https://369afa8d72cb4772b46936f7291218b7@kl.sentry.cloudlinux.com/2
APISERVER_SENTRY_DSN = https://3fc53cea3e0045aab42cc8e5358d6264@kl.sentry.cloudlinux.com/4

# provides additional info about specific service what kind of tests need to run
test:
	go test -v -coverprofile=coverage.out -coverpkg="github.com/kuberlogic/..." -run "$(RUN)" -timeout 60m -failfast $(ARGS) ./... ;\
	EXIT_CODE=$$?;\
	mkdir -p html/ ;\
	for pkg in apiserver operator; do \
		echo "mode: set" > $${pkg}.out; \
		cat coverage.out | grep "github.com/kuberlogic/$${pkg}" >> $${pkg}.out; \
		go tool cover -html=$${pkg}.out -o=html/$${pkg}.html; \
	done; \
	exit $$EXIT_CODE

build-minio:
	docker build minio/ -t $(IMG) --build-arg MINIO_ACCESS_KEY=$(MINIO_ACCESS_KEY) --build-arg MINIO_SECRET_KEY=$(MINIO_SECRET_KEY)

push-minio:
	docker push $(IMG)

deploy-minio:
	kubectl apply -f minio/deployment.yaml

endpoint-minio:
	@echo $(MINIO_ENDPOINT)

destroy-minio:
	kubectl delete -f minio/deployment.yaml

create-bucket:
	[ -f "./mc" ]  || wget https://dl.min.io/client/mc/release/linux-amd64/mc -O mc && chmod +x mc; \
	./mc alias set minio $(MINIO_ENDPOINT) $(MINIO_ACCESS_KEY) $(MINIO_SECRET_KEY); \
	./mc mb minio/$(BUCKET); \

show-minio-credentials:
	@echo "AWS_ACCESS_KEY_ID:" $(MINIO_ACCESS_KEY)
	@echo "AWS_SECRET_ACCESS_KEY:" $(MINIO_SECRET_KEY)
	@echo "Endpoint:" $(MINIO_ENDPOINT)
	@echo "Default bucket:" $(BUCKET)
