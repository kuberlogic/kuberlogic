#!/bin/bash

set -o pipefail

REGEX='(mysql|postgresql)$';
if [[ "$1" =~ $REGEX ]]; then
  echo $BASH_REMATCH;
fi