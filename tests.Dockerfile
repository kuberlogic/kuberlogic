FROM golang:1.16 as builder

# Copy the operator sources
WORKDIR /workspace/operator/
COPY modules/operator ./

WORKDIR /workspace/apiserver

# Copy the Go Modules manifests
COPY modules/apiserver/go.mod go.mod
COPY modules/apiserver/go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY modules/apiserver/cmd/ cmd/
COPY modules/apiserver/internal/ internal/
COPY modules/apiserver/util/ util/
COPY modules/apiserver/main.go main.go
COPY modules/apiserver/tests/ tests/

# Copy kubectl
RUN curl https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl -o /usr/bin/kubectl
RUN chmod +x /usr/bin/kubectl

# Build
WORKDIR /workspace/apiserver/tests
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go test -c -o integration.tests

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
#FROM golang:1.16
WORKDIR /
COPY --from=builder /workspace/apiserver/tests/integration.tests .
COPY --from=builder /usr/bin/kubectl /usr/bin/kubectl
USER nonroot:nonroot

ENTRYPOINT ["/integration.tests"]