# shellcheck shell=bash
# show cluster context at login

if [[ -z "${OCMC_ENGINE}" ]]
then
  # Not launched by ocm-container binary; skip automation
  # and allow the user to decide when/if this is necessary
  return
fi

if [ -n "$CLUSTER_ID" ] && [ -z "$SKIP_CLUSTER_CONTEXT" ]; then
	echo "Checking the context on $CLUSTER_ID"
	osdctl cluster context "$CLUSTER_ID"
fi
