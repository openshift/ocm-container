#!/usr/bin/env bash

set -euo pipefail

GITHUB_TOKEN="${GITHUB_TOKEN:-}"

if [[ -z "${GITHUB_TOKEN}" ]]; then
    echo "GITHUB_TOKEN is not set. Builds may be subject to GitHub rate limits."
fi

# NOTE: GITHUB_TOKEN does not need to be passed to the `make` command directly
# as it is already set in the environment. The Makefile will use it if it exists.
# Leaving it out here helps to potentially avoid issues with the token being exposed.

# Cleanup any existing manifests
make remove-manifests

# Build the images
# arm64 is not supported at this time, so we only build amd64
# Multiple architectures can be specified at once by passing a comma-separated list
# Example: ARCH="amd64,arm64"

ARCH="amd64"
echo "Building micro image for architecture: ${ARCH}"
make ARCHITECTURE=${ARCH} build-micro
make ARCHITECTURE=${ARCH} tag-micro

# Allow cache for minimal and full
# Full cache is invalidated from the micro build
# and this will allow us to re-use layers from these builds
# without unintentionally caching previous builds
echo "Building minimal image for architecture: ${ARCH}"
make CACHE="" build-minimal
make ARCHITECTURE=${ARCH} tag-minimal

echo "Building full image for architecture: ${ARCH}"
make CACHE="" build-full
make ARCHITECTURE=${ARCH} tag-full

make registry-login

# Push the images
for ARCH in amd64; do
    echo "Pushing images for architecture: ${ARCH}"
    make ARCHITECTURE=${ARCH} push-all
done

# Skipping the manifest steps as we're not currently building arm64 images
make push-manifests
