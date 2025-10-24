# shellcheck shell=bash

# oc config current-context will return a 1 for newly-opened ocm-container
# This prevents another attempt at login if using a terminal multiplexer
if ! oc config current-context &>/dev/null && [ -n "$INITIAL_CLUSTER_LOGIN" ]
then
  sre-login "$INITIAL_CLUSTER_LOGIN"
  cluster_info_env_export
fi

