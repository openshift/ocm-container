#!/usr/bin/env bash

# show cluster context at login

if [ -n "$CLUSTER_ID" -a -z "$SKIP_CLUSTER_CONTEXT" ]; then
	echo "Checking the context on $CLUSTER_ID"
	osdctl cluster context "$CLUSTER_ID"
fi
