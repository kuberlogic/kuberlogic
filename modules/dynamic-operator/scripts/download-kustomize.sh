#!/bin/bash

#
# CloudLinux Software Inc 2019-2021 All Rights Reserved
#

set -eou pipefail

function get_system {
  echo $(uname -s | tr '[:upper:]' '[:lower:]')
}

function get_arch {
  case "$(uname -m)" in
  x86_64 | amd64)
    echo amd64
    ;;
  i?86)
    echo s390x
    ;;
  arm*)
    echo arm64
    ;;
  powerpc | ppc64)
    echo ppc64le
    ;;
  *)
    echo "$(uname -m) is not supported"
    exit 1
    ;;
  esac
}

function main {
  version=$1
  target=$2
  if [ ! -f "$target" ]; then
    parentdir=$(dirname $target)

    if [ ! -d "$parentdir" ]; then
      echo "$parentdir"
      mkdir -p $parentdir
    fi

    tmpdir=$(mktemp -d)
    archive="$tmpdir/kustomize.tar.gz"
    url="https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F${version}/kustomize_${version}_$(get_system)_$(get_arch).tar.gz"
    echo "Downloading kustomize $1"
    curl --output $archive --location $url


    echo "Extracting archive to $2"
    tar -xvf $archive -C $parentdir

    rm -rf $tmpdir
    echo "Done."
  else
    echo "$(basename $target) file already exists"
  fi
}

[ $# -eq 0 ] && {
  echo "Usage: $0 <version> <target>"
  exit 1
}

main $@
