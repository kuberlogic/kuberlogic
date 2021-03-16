IMG=quay.io/kuberlogic/kuberlogic-agent
VERSION=0.1.0
SVC=""

build:
	stat Dockerfile-$(SVC)
	docker build -t $(IMG):$(VERSION)-$(SVC) -f Dockerfile-$(SVC) .

run:
	docker run -e "HOST=192.168.239.4" test-agent-client

release:
	docker push $(IMG):$(VERSION)

br: build release