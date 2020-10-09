#!/bin/bash -e

if [ "$I_AM_IN_CONTAINER" != "I-am-in-container" ]; then
  echo "must be run in container";
  exit 1;
fi

echo "in container";

#export veleroversion=v1.5.1

pushd /usr/local/;
velerofolder=velero-${veleroversion}-linux-amd64
velerotarfile=${velerofolder}.tar.gz
wget -q https://github.com/vmware-tanzu/velero/releases/download/${veleroversion}/${velerotarfile};
tar xzvf ${velerotarfile};
rm ${velerotarfile};
mv velero{-${veleroversion}-linux-amd64,};
ln -s /usr/local/velero/velero /usr/local/bin/velero;
velero completion bash > /etc/bash_completion.d/velero;
popd;
