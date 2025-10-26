# shellcheck shell=bash

# oc config current-context will return a 1 for newly-opened ocm-container
# This prevents another attempt at login if using a terminal multiplexer
if ! oc config current-context &>/dev/null && [ -z "$SKIP_CLUSTER_LOGIN" ] && [ -n "$CLUSTER_ID" ]
then
  sre-login "$CLUSTER_ID"
fi
