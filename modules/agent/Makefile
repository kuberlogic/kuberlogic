IMG=gitlab.corp.cloudlinux.com:5001/cloudmanaged/cloudmanaged/kuberlogic-agent
VERSION=0.0.8-3

build:
	docker build -t $(IMG):$(VERSION) .

run:
	docker run -e "HOST=192.168.239.4" test-agent-client

release:
	docker push $(IMG):$(VERSION)

br: build release