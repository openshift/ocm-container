#!/bin/bash -e

if [ "$I_AM_IN_CONTAINER" != "I-am-in-container" ]; then
  echo "must be run in container";
  exit 1;
fi

echo "in container";

mkdir -p ${HOME}/utils
mkdir -p /etc/profile.d
echo "export PATH=${PATH}:${HOME}/utils" > /etc/profile.d/localbin.sh
chmod +x /etc/profile.d/localbin.sh

mv /container-setup/utils/* ${HOME}/utils

# Cleanup Home Dir
rm /root/anaconda*
rm /root/original-ks.cfg
