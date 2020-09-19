#!/bin/bash -e

if [ "$1" != "I-am-in-container" ]; then
  echo "must be run in container";
  exit 1;
fi

source util.sh
echo "in container";

#export moactlversion=v0.0.5

mkdir /usr/local/moactl;
pushd /usr/local/moactl;
remove_coloring go get -v -u github.com/openshift/moactl;
ln -s /root/go/bin/moactl /usr/local/bin/moactl;
moactl completion bash >  /etc/bash_completion.d/moactl
popd;
