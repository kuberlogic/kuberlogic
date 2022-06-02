# Build the manager binary
FROM golang:1.16 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY cfg/ cfg/
COPY plugin/commons plugin/commons

# Build
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -a -o manager main.go

# Build plugin
COPY plugins/docker-compose plugins/docker-compose
RUN cd plugins/docker-compose && go build -a -o docker-compose-plugin .

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
COPY --from=builder /workspace/plugins/docker-compose/docker-compose-plugin .
USER 65532:65532

ENTRYPOINT ["/manager"]
