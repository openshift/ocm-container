#!/usr/bin/env bash

export OCM_URL=${OCM_URL:-production}

## Overwrite defaults with user-config
source /root/.config/ocm-container/env.source

## Extract mounted source CA's to /etc/pki/ca-trust/extracted
update-ca-trust
