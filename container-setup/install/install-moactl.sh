#!/bin/bash -e

if [ "$I_AM_IN_CONTAINER" != "I-am-in-container" ]; then
  echo "must be run in container";
  exit 1;
fi

source /container-setup/install/helpers.sh
echo "in container";

#export moactlversion=v0.0.5

pushd /usr/local;
# can be changed to git@github.com:openshift/moactl.git when ssh agent is passed to everyone with ease
git clone https://github.com/openshift/moactl.git;
pushd moactl;
make install;
ln -s /root/go/bin/moactl /usr/local/bin/moactl;
moactl completion bash >  /etc/bash_completion.d/moactl
popd;
popd;
