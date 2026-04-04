# shellcheck shell=bash

# oc config current-context will return a 1 for newly-opened ocm-container
# This prevents another attempt at login if using a terminal multiplexer
if ! oc config current-context &>/dev/null && [ -z "$SKIP_CLUSTER_LOGIN" ] && [ -n "$CLUSTER_ID" ]
then
  if [ -f "${HOME}/.session/sre-login-attempted" ]; then
    echo "Skipping automatic cluster login (previous attempt detected). Run 'sre-login $CLUSTER_ID' to retry manually."
  else
    mkdir -p "${HOME}/.session"
    touch "${HOME}/.session/sre-login-attempted"
    sre-login "$CLUSTER_ID"
  fi
fi
