#!/bin/bash -e

if [ "$I_AM_IN_CONTAINER" != "I-am-in-container" ]; then
  echo "must be run in container";
  exit 1;
fi

echo "in container";

echo "Installing moactl"
./install/install-moactl.sh

echo "Installing OCM"
./install/install-ocm.sh

echo "Installing oc"
./install/install-oc.sh

echo "Installing aws"
./install/install-aws.sh

echo "Installing KubePS1"
./install/install-kube_ps1.sh

echo "Installing osdctl"
./install/install-osdctl.sh

echo "Installing Velero"
./install/install-velero.sh

echo "Installing Utilities"
./install/install-utils.sh

cat /container-setup/install/bashrc_supplement.sh >> ~/.bashrc

rm -rf /container-setup
