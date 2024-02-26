#!/usr/bin/env bash

# Build the images
./.ci/build.sh

make TAG=latest-amd64 ARCHITECTURE=amd64 tag
make TAG=latest-arm64 ARCHITECTURE=arm64 tag
