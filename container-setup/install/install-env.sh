#!/bin/bash -e

# For aws
install_aws() {
#export awsclient=awscli-exe-linux-x86_64.zip
mkdir /usr/local/aws;
pushd /usr/local/aws;
wget -q https://awscli.amazonaws.com/${awsclient}
unzip ${awsclient}
rm ${awsclient}
./aws/install;
popd;
}

# For cluster login
install_cluster_login() {
pip3 install requests-html
pyppeteer-install
}

# For kub_ps1
install_kube_ps1() {
mkdir /usr/local/kube_ps1;
pushd /usr/local/kube_ps1;
wget -q https://raw.githubusercontent.com/drewandersonnz/kube-ps1/master/kube-ps1.sh;
popd;
}

# For oc
install_oc() {
#export osv4client=openshift-client-linux-4.3.5.tar.gz

# https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/
mkdir /usr/local/oc;
pushd /usr/local/oc;
wget -q https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/${osv4client};
tar xzvf ${osv4client};
rm ${osv4client};
ln -s /usr/local/oc/oc /usr/local/bin/oc;
oc completion bash >  /etc/bash_completion.d/oc
popd;
}

# For ocm
install_ocm() {
source /container-setup/install/helpers.sh

LATEST_RELEASE="https://github.com/openshift-online/ocm-cli/releases/latest/download/ocm-linux-amd64"

mkdir /usr/local/ocm \
  && pushd /usr/local/ocm

echo "Downloading the latest release"
curl -sSL ${LATEST_RELEASE} -o ocm-linux-amd64
echo "Validating binary against SHA256 sum"
curl -sSL ${LATEST_RELEASE}.sha256 | sha256sum --check --status
echo "Making binary executable"
chmod +x ocm-linux-amd64
echo "Symlinking binary to \"ocm\""
ln -s /usr/local/ocm/ocm-linux-amd64 /usr/local/bin/ocm
echo "Checking binary execution"
ocm version

popd

ocm completion > /etc/bash_completion.d/ocm
}

# For osdctl
install_osdctl() {
#export osdctlversion=v0.1.0

mkdir /usr/local/osdctl;
pushd /usr/local/osdctl;
wget -q https://github.com/openshift/osd-utils-cli/releases/download/${osdctlversion}/osdctl-linux-${osdctlversion};
mv osdctl{-linux-${osdctlversion},}
chmod +x osdctl
ln -s /usr/local/osdctl/osdctl /usr/local/bin/osdctl;
osdctl completion bash > /etc/bash_completion.d/osdctl;
popd;
}

# For rosa
install_rosa() {
pushd /usr/local;
# can be changed to git@github.com:openshift/moactl.git when ssh agent is passed to everyone with ease
git clone https://github.com/openshift/moactl.git;
mv moactl rosa;
pushd rosa;

# harden the moactl to use the latest tag and not master
# to override remove the following lines
LATEST_TAG=$(git describe --tags);
git checkout ${LATEST_TAG};

make install;
ln -s /root/go/bin/rosa /usr/local/bin/rosa;
rosa completion bash >  /etc/bash_completion.d/rosa
popd;
popd;
}

# For velero
install_velero() {
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
}

set -o errexit

# Check in container or not
if [ "$I_AM_IN_CONTAINER" != "I-am-in-container" ]; then
  echo "must be run in container";
  exit 1;
fi

echo "in container";

# Script usage.
if [ $# -eq 0 ]
then
  echo "Usage: `basename $0` [aws|cluster-login|kube|oc|ocm|osdctl|rosa|velero]"
fi

# Option
case "$1" in
	aws)
		install_aws
		echo "$1"
		;;
	cluster-login)
		install_cluster_login
		echo "$1"
		;;
	kube)
		install_kube_ps1
		echo "$1"
		;;
	oc)
		install_oc
		echo "$1"
		;;
	ocm)
		install_ocm
		echo "$1"
		;;
	osdctl)
		install_osdctl
		echo "$1"
		;;
	rosa)
		install_rosa
		echo "$1"
		;;
	velero)
		install_velero
		echo "$1"
		;;
	*)
		echo "Invaild option."
		;;
esac
