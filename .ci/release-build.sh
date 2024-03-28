#!/usr/bin/env bash

set -euo pipefail

# Build the images
./.ci/build.sh

make TAG=latest-amd64 ARCHITECTURE=amd64 tag
# make TAG=latest-arm64 ARCHITECTURE=arm64 tag

make TAG=latest-amd64 ARCHITECTURE=amd64 push
# make TAG=latest-arm64 ARCHITECTURE=arm64 push

make build-manifest
make push-manifest
