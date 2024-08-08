#!/usr/bin/env bash

if [[ -z "${OCMC_ENGINE}" ]]
then
  # Not launched by ocm-container binary; skip automation
  # and allow the user to decide when/if this is necessary
  return
fi

if ! ocm whoami &> /dev/null
then
  ocm login --url="${OCMC_OCM_URL:-https://api.openshift.com}" --use-device-code
else
  ocm config set url "${OCMC_OCM_URL}"
fi
