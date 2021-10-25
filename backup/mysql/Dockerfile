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
        wget \
        # mydumper dependencies
        libssl1.0.0 \
        libglib2.0-dev \
    && pip3 install --no-cache-dir awscli --upgrade \
    && wget https://github.com/maxbube/mydumper/releases/download/v0.9.5/mydumper_0.9.5-2.xenial_amd64.deb -O /tmp/installer.dpkg \
    && dpkg -i /tmp/installer.dpkg \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && rm /tmp/installer.dpkg \
    && curl -SL --progress-bar https://github.com/getsentry/sentry-cli/releases/download/1.70.1/sentry-cli-Linux-x86_64 > /usr/local/bin/sentry-cli \
    && chmod +x /usr/local/bin/sentry-cli

COPY dump.sh ./
RUN chmod +x ./*.sh

ENTRYPOINT ["/dump.sh"]
