#!/usr/bin/env bash

set -euo pipefail

GITHUB_TOKEN="${GITHUB_TOKEN:-}"

# Build the images
make GITHUB_TOKEN=${GITHUB_TOKEN} build-image-amd64
# make build-image-arm64

make ARCHITECTURE=amd64 tag
# make ARCHITECTURE=arm64 tag

make registry-login

make TAG=latest-amd64 ARCHITECTURE=amd64 push
# make TAG=latest-arm64 ARCHITECTURE=arm64 push

make build-manifest
make push-manifest
