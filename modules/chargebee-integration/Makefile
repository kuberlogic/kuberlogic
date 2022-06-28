.EXPORT_ALL_VARIABLES:

CHARGEBEE_SITE = imunify360-test
CHARGEBEE_KEY = test_aTwMcu4LDMU3eCamXCJYcubL60tGDYbGeU
KUBERLOGIC_APISERVER_SCHEME = http
KUBERLOGIC_APISERVER_HOST = 198.51.100.0
KUBERLOGIC_APISERVER_TOKEN = 8ZTjsD3t2Q3Yq-C4-hoahcFn
KUBERLOGIC_DOMAIN = example.com
KUBERLOGIC_TYPE = docker-compose

.PHONY: run
run:
	go run main.go

.PHONY: test
test:
	go test -v ./... -coverprofile cover.out