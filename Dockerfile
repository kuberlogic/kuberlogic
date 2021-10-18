# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /

ARG BIN
COPY ${BIN} /cmd
USER nonroot:nonroot

ENTRYPOINT ["/cmd"]
