# shellcheck shell=bash
test -f /usr/local/bin/velero || ln -s ~/.local/bin/install-velero /usr/local/bin/velero

### Set the default namespace for velero to avoid using
### --namespace=openshift-velero each time
if [ -n "$DEFAULT_VELERO_NS" ]
then
  velero client config set namespace="${DEFAULT_VELERO_NS}"
fi
