#!/bin/bash
set -euo pipefail

platform="unknown"
tag="latest"

main() {
  case $(uname -s) in
  Linux) platform="linux" ;;
  Darwin) platform="darwin" ;;
  *)
    echo "Unsupported OS"
    exit 1
    ;;
  esac

  echo "Detecting os: " $platform

  curl -L -o /usr/local/bin/kuberlogic-installer \
    https://github.com/kuberlogic/kuberlogic/releases/download/$tag/kuberlogic-installer-$platform-amd64
  chmod +x /usr/local/bin/kuberlogic-installer
  /usr/local/bin/kuberlogic-installer install all
}

main "$@"
