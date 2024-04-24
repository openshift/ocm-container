#!/usr/bin/env bash

export OCM_URL=${OCM_URL:-production}

## Extract mounted source CA's to /etc/pki/ca-trust/extracted
update-ca-trust
