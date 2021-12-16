#!/usr/bin/env bash

set -eEuo pipefail

### Set the default namespace for velero to avoid using
### --namespace=openshift-velero each time
export DEFAULT_VELERO_NS=4{DEFAULT_VELERO_NS:-}
if [ -n "$DEFAULT_VELERO_NS" ]
then
  velero client config set namespace=${DEFAULT_VELERO_NS}
fi

set +eEuo pipefail
