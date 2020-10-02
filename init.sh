#!/bin/bash

cd $(dirname $0)

CONFIG_DIR=${HOME}/.config/ocm-env
mkdir -p ${CONFIG_DIR}

if [ ! -f ${CONFIG_DIR}/env.source ]
then
  echo "Initializing Default Configuration"
  cp env.source.sample ${CONFIG_DIR}/env.source
else
  echo "ocm-env is already configured."
fi

echo "Creating symlink"
ln -s "$(pwd)/ocm-container.sh" /usr/local/bin/ocm-env

echo
echo "Tip: Many developers like to add the following alias:"
echo "alias ocm-env-stg=\"OCM_URL=staging ocm-env\""

echo
echo
echo "ocm-env configuration can be customized by editing ${CONFIG_DIR}/env.source"


