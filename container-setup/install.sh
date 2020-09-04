#!/bin/bash -e

if [ "$1" != "I-am-in-container" ]; then
  echo "must be run in container";
  exit 1;
fi

echo "in container";

yum -y install \
    bash-completion \
    findutils \
    git \
    golang \
    jq \
    make \
    procps-ng \
    rsync \
    sshuttle \
    vim-enhanced \
    wget;

yum clean all;

export moactlversion=v0.0.5
./install-moactl.sh $1

./install-ocm.sh $1

#export osv4client=openshift-client-linux-4.3.5.tar.gz
./install-oc.sh $1

export awsclient=awscli-exe-linux-x86_64.zip
./install-aws.sh $1

./install-kube_ps1.sh $1

export osdctlversion=v0.1.0
./install-osdctl.sh $1

# Activate all environment variables from env.source
echo 'source /root/env.source' >> ~/.bashrc

echo 'source /container-setup/bashrc_supplement.sh' >> ~/.bashrc
