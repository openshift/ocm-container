#!/bin/bash -e

if [ "$I_AM_IN_CONTAINER" != "I-am-in-container" ]; then
  echo "must be run in container";
  exit 1;
fi

echo "in container";

mkdir /usr/local/kube_ps1;
pushd /usr/local/kube_ps1;
wget -q https://raw.githubusercontent.com/drewandersonnz/kube-ps1/master/kube-ps1.sh;
popd;
