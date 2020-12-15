#!/bin/bash -e

if [ "$I_AM_IN_CONTAINER" != "I-am-in-container" ]; then
  echo "must be run in container";
  exit 1;
fi

echo "in container";

mkdir -p ${HOME}/utils

mv /container-setup/utils/bin/* ${HOME}/utils

# Cleanup Home Dir
rm /root/anaconda*
rm /root/original-ks.cfg

# bashrc supplement
cat /container-setup/utils/bashrc_supplement.sh >> ${HOME}/.bashrc
