IMG=quay.io/kuberlogic/kl-ui
TAG=latest

build:
	docker build -t $(IMG):$(TAG) .

release: build
	docker push $(IMG):$(TAG)
