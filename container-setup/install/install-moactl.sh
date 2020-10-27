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

# harden the moactl to use the latest tag and not master
# to override remove the following lines
LATEST_TAG=$(git describe --tags);
git checkout ${LATEST_TAG};

make install;
ln -s /root/go/bin/rosa /usr/local/bin/rosa;
rosa completion bash >  /etc/bash_completion.d/rosa
popd;
popd;
