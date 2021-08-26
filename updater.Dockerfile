FROM golang:1.16 as builder

# Copy the operator sources
WORKDIR /workspace/operator/
COPY modules/operator ./

WORKDIR /workspace/updater

# Copy the Go Modules manifests
COPY modules/updater/go.mod go.mod
COPY modules/updater/go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY modules/updater/main.go main.go

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o updater main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/updater .
USER nonroot:nonroot

ENTRYPOINT ["/updater"]
