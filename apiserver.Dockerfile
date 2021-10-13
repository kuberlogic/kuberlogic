FROM golang:1.16 as builder

# Copy the operator sources
WORKDIR /workspace/operator/
COPY modules/operator ./

# Copy & build the apiserver
WORKDIR /workspace/apiserver/

# Copy the Go Modules manifests
COPY modules/apiserver/go.mod go.mod
COPY modules/apiserver/go.sum go.sum

# Cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY modules/apiserver/cmd cmd/
COPY modules/apiserver/internal internal/
COPY modules/apiserver/util util/
COPY modules/apiserver/main.go main.go

ARG VERSION
ARG REVISION
ARG BUILD_TIME

# Build
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GO111MODULE=on \
    go build \
    -ldflags " \
    -X github.com/kuberlogic/kuberlogic/modules/apiserver/cmd.sha1ver=$REVISION \
    -X github.com/kuberlogic/kuberlogic/modules/apiserver/cmd.buildTime=$BUILD_TIME \
    -X github.com/kuberlogic/kuberlogic/modules/apiserver/cmd.ver=$VERSION"  \
    -a -o apiserver main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/apiserver .
USER nonroot:nonroot

ENTRYPOINT ["/apiserver"]
