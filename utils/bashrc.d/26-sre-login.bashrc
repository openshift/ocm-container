# shellcheck shell=bash

if [[ -z "${OCMC_ENGINE}" ]]
then
  # Not launched by ocm-container binary; skip automation
  # and allow the user to decide when/if this is necessary
  return
fi

# oc config current-context will return a 1 for newly-opened ocm-container
# This prevents another attempt at login if using a terminal multiplexer
if ! oc config current-context &>/dev/null && [ -n "$INITIAL_CLUSTER_LOGIN" ]
then
  sre-login "$INITIAL_CLUSTER_LOGIN"
  cluster_info_env_export
fi

