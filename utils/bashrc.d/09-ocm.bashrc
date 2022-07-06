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

# Wrap the ocm backplane console command to handle automation for
# port mapping inside the container
ocm() {
  if [[ "$@" =~ "backplane console" ]]; then
    shift 2
    echo "/root/.local/bin/cluster-console $@"
    command /root/.local/bin/cluster-console
  else
    command ocm "$@"
  fi
}
