#!/bin/bash -e

if [ "$I_AM_IN_CONTAINER" != "I-am-in-container" ]; then
  echo "must be run in container";
  exit 1;
fi

echo "in container";

echo "Installing moactl"
./install-moactl.sh

echo "Installing OCM"
./install-ocm.sh

echo "Installing oc"
./install-oc.sh

echo "Installing aws"
./install-aws.sh

echo "Installing KubePS1"
./install-kube_ps1.sh

echo "Installing osdctl"
./install-osdctl.sh

echo "Installing Velero"
./install-velero.sh

# Activate all environment variables from env.source
echo 'source /root/env.source' >> ~/.bashrc

echo 'source /container-setup/bashrc_supplement.sh' >> ~/.bashrc
