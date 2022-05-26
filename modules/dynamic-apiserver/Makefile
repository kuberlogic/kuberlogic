
.PHONY: recreate-generated-server
recreate-generated-server:
	rm -r internal/generated/models internal/generated/restapi
	mkdir -p internal/generated

.PHONY: recreate-generated-client
recreate-generated-client:
	rm -r internal/generated/client
	mkdir -p internal/generated

.PHONY: generate-server
generate-server: recreate-generated-server
	swagger generate server \
		--spec openapi.yaml \
		--target internal/generated \
		--config-file swagger-templates/default-server.yml \
		--template-dir swagger-templates/templates/ \
		--name kuberlogic

.PHONY: generate-client
generate-client: recreate-generated-client
	swagger generate client \
		--spec openapi.yaml \
		--target internal/generated/

.PHONY: test
test:
	go test -v ./... -coverprofile cover.out

.PHONY: build-cli
build-cli:
	go build -o kuberlogic cmd/cli/main.go

.PHONY: run
run:
	go run main.go

.PHONY: generate
generate: generate-server generate-client