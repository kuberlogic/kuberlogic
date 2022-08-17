#!/bin/bash
set -e

echo "Getting latest released Kuberlogic version..."
TAG=$(curl -s -H "Accept: application/vnd.github+json" https://api.github.com/repos/kuberlogic/kuberlogic/releases/latest | awk -F\" '/tag_name/ {print $4}')
OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

INSTALL_DIR=/usr/local/bin
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
  INSTALL_DIR=/bin
fi

TMPFILE=$(date +%s)

echo "Installing Kuberlogic version ${TAG}..."
wget -q "https://github.com/kuberlogic/kuberlogic/releases/download/${TAG}/kuberlogic_${ARCH}_${OS}" -O "${TMPFILE}"
install -b "${TMPFILE}" $INSTALL_DIR/kuberlogic
rm "${TMPFILE}"

echo "KuberLogic CLI installed at ${INSTALL_DIR}/kuberlogic\nUse 'kuberlogic help' command to start."