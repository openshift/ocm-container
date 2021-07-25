#!/usr/bin/env bash

if [ "x${OFFLINE_ACCESS_TOKEN}" == "x" ]; then
  echo "FAILURE: must set env variable OFFLINE_ACCESS_TOKEN";
  exit 1;
fi

if [ "$OCM_URL" == "" ];
then
  OCM_URL="https://api.openshift.com"
fi

CLI="${CLI:-ocm}"
if [[ "${CLI}" == "ocm" ]]; then
  LOGIN_ENV='--url';
elif [[ "${CLI}" == "moactl" ]]; then
  LOGIN_ENV='--env';
fi

"${CLI}" login --token=$OFFLINE_ACCESS_TOKEN ${LOGIN_ENV}=$OCM_URL
