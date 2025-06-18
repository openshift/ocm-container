#!/usr/bin/env bash

set -euo pipefail

GITHUB_TOKEN="${GITHUB_TOKEN:-}"

if [[ -z ${GITHUB_TOKEN} ]]
then
  echo "No GITHUB_TOKEN set; downloads may be rate-limited"
else
  echo "GITHUB_TOKEN set"
fi

# Build the images
make BUILD_ARGS="--no-cache --build-arg GITHUB_TOKEN=${GITHUB_TOKEN}" build-image-amd64
make TAG=latest-amd64 ARCHITECTURE=amd64 tag

# ARM builds not currently supported in CI
# make BUILD_ARGS="--no-cache --build-arg GITHUB_TOKEN=${GITHUB_TOKEN}"  build-image-arm64
# make TAG=latest-arm64 ARCHITECTURE=arm64 tag
