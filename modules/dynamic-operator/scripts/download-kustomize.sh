#!/bin/bash

#
# CloudLinux Software Inc 2019-2021 All Rights Reserved
#

set -e -o pipefail -u

function get_system {
  echo $(uname -s | tr '[:upper:]' '[:lower:]')
}

function get_arch {
  case "$(uname -m)" in
  x86_64|amd64)
    echo amd64;;
  i?86)
    echo s390x;;
  arm*)
    echo arm64;;
  powerpc|ppc64)
    echo ppc64le;;
  *)
    echo "$(uname -m) is not supported" > /dev/tty;
    exit 1;;
esac
}

function main {
  version=$1;
  target=$2;
  if [ ! -f "$target" ]; then
    tmpdir=$(mktemp -d);
    archive="$tmpdir/kustomize.tar.gz"
    url="https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F${version}/kustomize_${version}_$(get_system)_$(get_arch).tar.gz"
    echo "Downloading kustomize $1" > /dev/tty
    curl -o $archive -L $url
    mkdir $(dirname $target)
    echo "Extracting archive to $2" > /dev/tty
    tar -xvf $archive -C $(dirname $target)
    rm -rf $tmpdir;
    echo "Done."
  else
    echo "Directory '$target' already exists"
  fi
}

[ $# -eq 0 ] && { echo "Usage: $0 <version> <target>"; exit 1; }
main $@