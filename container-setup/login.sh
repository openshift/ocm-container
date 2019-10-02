#!/bin/bash

if [ "x${OFFLINE_ACCESS_TOKEN}" == "x" ]; then
  echo "must set OFFLINE_ACCESS_TOKEN";
  exit 1;
fi

TOKEN=$(curl \
--silent \
--data-urlencode "grant_type=refresh_token" \
--data-urlencode "client_id=cloud-services" \
--data-urlencode "refresh_token=${OFFLINE_ACCESS_TOKEN}" \
https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token | \
jq -r .access_token)

ocm login --token=$TOKEN

if [ "$1" != "" ];
then
    oc logout 2>/dev/null
    ocm cluster login $1
else
    ocm cluster list --managed
    exit 1
fi
