#!/bin/bash

CLUSTERID=$1

CLI="${CLI:-ocm}"
if [[ "${CLI}" == "ocm" ]]; then
  OCM_LIST_ADDITIONAL_ARG='--managed --columns id,name,api.url,openshift_version,region.id,state,external_id';
  LOGIN_ENV='--url';
elif [[ "${CLI}" == "moactl" ]]; then
  LOGIN_ENV='--env';
fi

if [ "x${OFFLINE_ACCESS_TOKEN}" == "x" ]; then
  echo "FAILURE: must set env variable OFFLINE_ACCESS_TOKEN";
  exit 1;
fi

if [ "$OCM_URL" == "" ];
then
  OCM_URL="https://api.openshift.com"
fi

TOKEN=$(curl \
--silent \
--data-urlencode "grant_type=refresh_token" \
--data-urlencode "client_id=cloud-services" \
--data-urlencode "refresh_token=${OFFLINE_ACCESS_TOKEN}" \
https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token | \
jq -r .access_token)

"${CLI}" login --token=$TOKEN ${LOGIN_ENV}=$OCM_URL

if [ "${CLUSTERID}" != "" ]; then
    oc logout 2>/dev/null
    CLUSTER_DATA=$("${CLI}" describe cluster ${CLUSTERID} --json || exit )
    CLUSTER_API_TYPE=$( echo ${CLUSTER_DATA} | jq --raw-output .api.listening )
    CONSOLE_URL=$(echo ${CLUSTER_DATA} | jq --raw-output .console.url)
    if [[ ${CLUSTER_API_TYPE} == "internal" ]]
    then
            echo "FAILURE: cannot connect to a private cluster"
            exit 1
    fi

    "${CLI}" cluster login ${CLUSTERID} --username ${OCM_USER} --console >/dev/null 2>&1 \
        || (echo "FAILURE: unable to login, Try accessing '${CONSOLE_URL}'" && exit 1)
else
    "${CLI}" list cluster ${OCM_LIST_ADDITIONAL_ARG}
    exit 1 # exit the container
fi
