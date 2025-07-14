#!/usr/bin/env bash

set -euo pipefail

GITHUB_TOKEN="${GITHUB_TOKEN:-}"

# Build the images
make BUILD_ARGS="--build-arg GITHUB_TOKEN=${GITHUB_TOKEN}" build-image-amd64
# make BUILD_ARGS="--no-cache" build-image-arm64