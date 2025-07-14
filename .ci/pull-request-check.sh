#!/usr/bin/env bash

set -euo pipefail

GITHUB_TOKEN="${GITHUB_TOKEN:-}"

if [[ -z "${GITHUB_TOKEN}" ]]; then
    echo "GITHUB_TOKEN is not set. Builds may be subject to GitHub rate limits."
fi

# NOTE: GITHUB_TOKEN does not need to be passed to the `make` command directly
# as it is already set in the environment. The Makefile will use it if it exists.
# Leaving it out here helps to potentially avoid issues with the token being exposed.

# The `make build` target will build the full image to validate
make build-full
