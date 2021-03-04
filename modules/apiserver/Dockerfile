FROM golang:1.13 as builder

WORKDIR /workspace

# create credetials for private repo
ADD build-data/ssh-config /root/.ssh/config
ADD build-data/git-config /root/.gitconfig
ADD build-data/.github.key /root/.ssh/id_rsa
RUN chmod 600 /root/.ssh/id_rsa

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# remove credentials for private repo
RUN rm -rf /root/.ssh/ /root/.gitconfig

# Copy the go source
COPY cmd/ cmd/
COPY internal/ internal/
COPY util/ util/
COPY main.go main.go

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o apiserver main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/apiserver .
USER nonroot:nonroot

ENTRYPOINT ["/apiserver"]
