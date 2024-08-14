# shellcheck shell=bash

function cluster_info_env_export(){
  SEARCH_STRING="id like '$INITIAL_CLUSTER_LOGIN' or external_id like '$INITIAL_CLUSTER_LOGIN' or name like '$INITIAL_CLUSTER_LOGIN' or display_name like '$INITIAL_CLUSTER_LOGIN'"
  cluster_details=$(ocm list clusters --parameter=search="(($SEARCH_STRING))" --columns "id, external_id, name" --no-headers)
  CLUSTER_ID=$(awk '{print $1}' <<< $cluster_details)
  CLUSTER_UUID=$(awk '{print $2}' <<< $cluster_details)
  CLUSTER_NAME=$(awk '{print $3}' <<< $cluster_details)
  export CLUSTER_ID CLUSTER_UUID CLUSTER_NAME
}

function cluster_function() {
  info="$(ocm backplane status 2> /dev/null)"
  if [ $? -ne 0 ]; then return; fi
  clustername=$(grep "Cluster Name" <<< $info | awk '{print $3}')
  baseid=$(grep "Cluster Basedomain" <<< $info | awk '{print $3}' | cut -d'.' -f1,2)
  echo $clustername.$baseid
}
