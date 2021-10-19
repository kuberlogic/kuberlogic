FROM ubuntu:18.04

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN apt-get update     \
    && apt-get install --no-install-recommends -y \
        apt-utils \
        ca-certificates \
        lsb-release \
        pigz \
        python3-pip \
        python3-setuptools \
        curl \
        jq \
        gnupg \
    && pip3 install --no-cache-dir awscli --upgrade \
    && echo "deb http://apt.postgresql.org/pub/repos/apt/ $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list \
    && cat /etc/apt/sources.list.d/pgdg.list \
    && curl --silent https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - \
    && apt-get update \
    && apt-get install --no-install-recommends -y  \
        postgresql-client-13  \
        postgresql-client-12  \
        postgresql-client-11  \
        postgresql-client-10  \
        postgresql-client-9.6 \
        postgresql-client-9.5 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && curl -sL https://sentry.io/get-cli/ | bash


COPY restore.sh pg_extract.sh recreatedb.sql /app/
RUN chmod +x /app/*.sh

WORKDIR /app/

ENV PG_DIR=/usr/lib/postgresql

ENTRYPOINT ["/app/restore.sh"]
