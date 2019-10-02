#!/bin/bash

if [ "$1" != "I-am-in-container" ]; then
  echo "must be run in container";
  exit 1;
fi

echo "in container";

yum -y install \
    bash-completion \
    findutils \
    git \
    jq \
    wget \
    vim-enhanced \
    golang;

yum clean all;

go get -u github.com/openshift-online/ocm-cli/cmd/ocm;
ln -s /root/go/bin/ocm /usr/local/bin/ocm;

# https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/
mkdir /usr/local/oc;
pushd /usr/local/oc;
wget -q https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/openshift-client-linux-4.1.18.tar.gz;
tar xzvf openshift-client-linux-4.1.18.tar.gz;
ln -s /usr/local/oc/oc /usr/local/bin/oc;
popd;
