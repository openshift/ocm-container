#!/usr/bin/env bash

if [ "$OCM_URL" == "" ]
then
  OCM_URL="https://api.openshift.com"
fi

if ! ocm whoami &> /dev/null
then
  ocm login --url=$OCM_URL --use-device-code
fi

# Wrap the ocm backplane console command to handle automation for
# port mapping inside the container
ocm() {
  if [[ "$@" =~ "backplane console" ]]
  then
    shift 2
    echo "/root/.local/bin/cluster-console $@"
    command /root/.local/bin/cluster-console
  else
    command ocm "$@"
  fi
}
