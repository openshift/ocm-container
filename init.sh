#!/bin/bash

cd $(dirname $0)
CONFIG_DIR=${HOME}/.config/ocm-container

function init_new_config() {
  if [ ! -f ${CONFIG_DIR}/env.source ]
  then
    echo "Initializing Default Configuration"
    cp env.source.sample ${CONFIG_DIR}/env.source
  else
    echo "ocm-container is already configured."
  fi
}

mkdir -p ${CONFIG_DIR}

if [ -f env.source ]
then
  echo "We see you already have a local configuration file.  Would you like us to:"
  echo "  1: Move the config file to the new config location"
  echo "  2: Symlink the config file to the new config location"
  echo "  3: Create a new config file at the new config location"
  echo "  4: Do nothing and exit"
  read -n 1 -p "Select 1-4: " prev_config_selection
  echo

  if [ $prev_config_selection == "1" ]
  then
    echo "Moving configuration file"
    mv env.source ${CONFIG_DIR}
  elif [ $prev_config_selection == "2" ]
  then
    echo "Symlinking configuration file"
    ln -s env.source $CONFIG_DIR/env.source
  elif [ $prev_config_selection == "3" ]
  then
    init_new_config
  else
    echo "Exiting."
    exit 0
  fi
else
  init_new_config
fi

echo "Creating symlink for ocm-container binary"
ln -sfn "$(pwd)/ocm-container.sh" /usr/local/bin/ocm-container

echo
echo "Tip: Many developers like to add the following alias:"
echo "alias ocm-container-stg=\"OCM_URL=staging ocm-container\""

echo
echo
echo "ocm-container configuration can be customized by editing ${CONFIG_DIR}/env.source"



