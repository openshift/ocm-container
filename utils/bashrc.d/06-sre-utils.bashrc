#!/usr/bin/env bash

set -eEuo pipefail

export INITIAL_CLUSTER_LOGIN=${INITIAL_CLUSTER_LOGIN:-}

if [ -n "$INITIAL_CLUSTER_LOGIN" ]
then
  sre-login $INITIAL_CLUSTER_LOGIN
fi

function cluster_function() {
  oc config view  --minify --output 'jsonpath={..server}' | cut -d. -f2-4
}

set +eEuo pipefail
