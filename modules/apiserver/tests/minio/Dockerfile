FROM minio/minio:latest

ARG MINIO_ACCESS_KEY
ARG MINIO_SECRET_KEY
ENV MINIO_ACCESS_KEY=$MINIO_ACCESS_KEY
ENV MINIO_SECRET_KEY=$MINIO_SECRET_KEY

RUN curl -sSL --tlsv1.2 \
         -O 'https://raw.githubusercontent.com/minio/kes/master/root.key' \
         -O 'https://raw.githubusercontent.com/minio/kes/master/root.cert'

ENV MINIO_KMS_KES_ENDPOINT=https://play.min.io:7373
ENV MINIO_KMS_KES_KEY_FILE=root.key
ENV MINIO_KMS_KES_CERT_FILE=root.cert
ENV MINIO_KMS_KES_KEY_NAME=my-minio-key

