
.PHONY: recreate-generated-directory
recreate-generated-directory:
	rm -r internal/generated
	mkdir -p internal/generated

.PHONY: generate
generate: recreate-generated-directory
	swagger generate server \
		--spec openapi.yaml \
		--target internal/generated \
		--config-file swagger-templates/default-server.yml \
		--template-dir swagger-templates/templates/ \
		--name kuberlogic

.PHONY: generate-cli
generate-cli:
	swagger generate cli \
		--spec openapi.yaml \
		--target internal/generated/ \
		--cli-app-name kuberlogic

.PHONY: test
test:
	go test -v ./... -coverprofile cover.out

build-cli:
	go build -o kuberlogic internal/generated/cmd/kuberlogic/main.go