#!/bin/bash -e

if [ "$I_AM_IN_CONTAINER" != "I-am-in-container" ]; then
  echo "must be run in container";
  exit 1;
fi

echo "in container";
source util.sh

remove_coloring go get -v -u github.com/openshift-online/ocm-cli/cmd/ocm;
ln -s /root/go/bin/ocm /usr/local/bin/ocm;
ocm completion bash > /etc/bash_completion.d/ocm
