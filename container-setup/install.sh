#!/bin/bash -e

if [ "$1" != "I-am-in-container" ]; then
  echo "must be run in container";
  exit 1;
fi

echo "in container";

#export osv4client=openshift-client-linux-4.3.5.tar.gz
export awsclient=awscli-exe-linux-x86_64.zip

yum -y install \
    bash-completion \
    findutils \
    git \
    golang \
    jq \
    make \
    vim-enhanced \
    rsync \
    wget;

yum clean all;

go get -u github.com/openshift-online/ocm-cli/cmd/ocm;
ln -s /root/go/bin/ocm /usr/local/bin/ocm;

# https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/
mkdir /usr/local/oc;
pushd /usr/local/oc;
wget -q https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/${osv4client};
tar xzvf ${osv4client};
rm ${osv4client};
ln -s /usr/local/oc/oc /usr/local/bin/oc;
popd;

mkdir /usr/local/aws;
pushd /usr/local/aws;
wget -q https://awscli.amazonaws.com/${awsclient}
unzip ${awsclient}
rm ${awsclient}
./aws/install;
popd;

mkdir /usr/local/kube_ps1;
pushd /usr/local/kube_ps1;
wget -q https://raw.githubusercontent.com/drewandersonnz/kube-ps1/master/kube-ps1.sh;
popd;

echo 'source /container-setup/bashrc_supplement.sh' >> ~/.bashrc
