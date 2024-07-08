#!/usr/bin/env bash

if [ "$OCM_URL" == "" ]
then
  OCM_URL="https://api.openshift.com"
fi

if ! ocm whoami &> /dev/null
then
  if [[ -n $OFFLINE_ACCESS_TOKEN ]] && grep -e "stage" -e "integration" <<< $OCM_URL &> /dev/null
  then
    ## Warn users trying to use offline access tokens in stage/int who have not passed through config
    echo "ERROR: Offline Access tokens will no longer work on staging or integration."
    echo "INFO: falling back to logging in with device code"
    echo "INFO: To prevent this from happening every time you create a container - log into OCM outside of the"
    echo "      container and ocm-container will pass the ocm config file through to the container"
    ocm login --url=$OCM_URL --use-device-code

  elif [[ -n $OFFLINE_ACCESS_TOKEN ]]
  then
    ## Warn users trying to use offline access tokens in prod, but allow it to continue working for now
    echo "WARN: Offline Access Tokens are being deprecated"
    echo "      Consider logging into ocm outside of the container with short lived access tokens"
    ocm login --url=$OCM_URL --token $OFFLINE_ACCESS_TOKEN
  fi
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
