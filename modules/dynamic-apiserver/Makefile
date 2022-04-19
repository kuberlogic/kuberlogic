
.PHONY: create-internal-directory
create-internal-directory:
	mkdir -p internal/generated

.PHONY: generate
generate: create-internal-directory
	swagger generate server \
		--spec openapi.yaml \
		--target internal/generated \
		--config-file swagger-templates/default-server.yml \
		--template-dir swagger-templates/templates/ \
		--name kuberlogic

.PHONY: test
test:
	go test -v ./... -coverprofile cover.out