#!/bin/bash

CLUSTERID=$1

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

ocm login --token=$TOKEN --url=$OCM_URL

if [ "${CLUSTERID}" != "" ];
then
    ac logout 2>/dev/null
    CLUSTER_DATA=$(ocm describe cluster ${CLUSTERID} --json || exit )
    CLUSTER_API_TYPE=$( echo ${CLUSTER_DATA} | jq --raw-output .api.listening )
    CONSOLE_URL=$(echo ${CLUSTER_DATA} | jq --raw-output .console.url)
    if [[ ${CLUSTER_API_TYPE} == "internal" ]]
    then
	    echo "FAILURE: cannot connect to a private cluster"
	    exit 1
    fi

    ocm cluster login ${CLUSTERID} --username ${OCM_USER} --console >/dev/null 2>&1 \
        || (echo "FAILURE: unable to login, Try accessing '${CONSOLE_URL}'" && exit 1)
else
    ocm list cluster --managed --columns id,name,api.url,openshift_version,region.id,state,external_id
    exit 1 # exit the container
fi
