# shellcheck shell=bash
# show cluster context at login

if [ -n "$CLUSTER_ID" ] && [ -z "$SKIP_CLUSTER_CONTEXT" ]; then
	echo "Checking the context on $CLUSTER_ID"
	osdctl cluster context --cluster-id "$CLUSTER_ID"
fi
