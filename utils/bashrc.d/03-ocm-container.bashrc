#!/usr/bin/env bash

set -eEuo pipefail

export OCM_URL=${OCM_URL:-production}

## Overwrite defaults with user-config
source /root/.config/ocm-container/env.source

set +eEuo pipefail
