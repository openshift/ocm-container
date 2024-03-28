#!/usr/bin/env bash

set -euo pipefail

# Build the images
./.ci/build.sh --no-cache

make TAG=latest-amd64 ARCHITECTURE=amd64 tag
make TAG=latest-arm64 ARCHITECTURE=arm64 tag
