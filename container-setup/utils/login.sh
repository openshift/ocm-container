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
    if [[ ${CONSOLE_URL} == "null" ]]
    then
        echo "FAILURE: console url is not yet exposed, please wait"
        exit 1
    fi

    if [[ ${CLUSTER_API_TYPE} == "internal" ]]
    then
        echo "INFO: requested ${CLUSTERID} is private"
        echo "attempting to tunnel connection across, prerequisites are located in https://github.com/openshift/ops-sop/blob/master/v4/howto/private-clusters.md"
        echo
        echo "$ ssh-add"
        echo "$ ssh-add -l"
        ssh-add -l
        echo
        echo "to tunnel your connection to the cluster, run:"
        echo "$ ocm tunnel '${CLUSTERID}' &"
    fi

    # '--browser' is a temporal flag and might change soon
    "${CLI}" cluster login ${CLUSTERID} --username ${OCM_USER} --token  \
        || (echo "FAILURE: unable to login" && echo "Use '${CONSOLE_URL}' to get the console or use Login URL for recieving the token" && exit 1)
else
    "${CLI}" list cluster ${OCM_LIST_ADDITIONAL_ARG}
    exit 1 # exit the container
fi
