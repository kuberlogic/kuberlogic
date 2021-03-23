IMG=quay.io/kuberlogic/kuberlogic-agent
VERSION=0.1.0
SVC=""

build:
	docker build -t $(IMG):$(VERSION)-$(SVC) -f client.Dockerfile .

release:
	docker push $(IMG):$(VERSION)

br: build release