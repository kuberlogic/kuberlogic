# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /

ARG BIN
ARG PLUGIN
COPY bin/manager /manager
COPY bin/docker-compose-plugin /docker-compose-plugin
USER 65532:65532

ENTRYPOINT ["/manager"]
