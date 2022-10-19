#!/usr/bin/env bash

cd "$(dirname "$0")" || exit
CONFIG_DIR=${HOME}/.config/ocm-container

if [[ -z ${aliases_file} ]]; then
	echo "run command with aliases file"
	echo 'aliases_file=$(alias) '"$0 " "$@"
	exit
fi

function init_new_config() {
  if [ ! -f "${CONFIG_DIR}/env.source" ]
  then
    echo "Initializing Default Configuration"
    cp env.source.sample "${CONFIG_DIR}/env.source"
  else
    echo "ocm-container is already configured."
  fi
}

mkdir -p "${CONFIG_DIR}"

if [ -f env.source ]
then
  echo "We see you already have a local configuration file.  Would you like us to:"
  echo "  1: Move the config file to the new config location"
  echo "  2: Symlink the config file to the new config location"
  echo "  3: Create a new config file at the new config location"
  echo "  4: Do nothing and exit"
  read -r -n 1 -p "Select 1-4: " prev_config_selection
  echo

  if [ "$prev_config_selection" == "1" ]
  then
    echo "Moving configuration file"
    mv env.source "${CONFIG_DIR}"
  elif [ "$prev_config_selection" == "2" ]
  then
    echo "Symlinking configuration file"
    ln -s env.source "$CONFIG_DIR/env.source"
  elif [ "$prev_config_selection" == "3" ]
  then
    init_new_config
  else
    echo "Exiting."
    exit 0
  fi
else
  init_new_config
fi

if ! [[ -L /usr/local/bin/ocm-container ]]; then 
  echo "Creating symlink for ocm-container binary (requires sudo permissions...)"
  sudo ln -sfn "$(pwd)/ocm-container.sh" /usr/local/bin/ocm-container
fi

# ! $( alias ocm-container-stg ) || 
if [[ $(echo "${aliases_file}" | grep -cw 'ocm-container-stg=') -eq 0 ]]; then
  echo
  echo "Tip: Many developers like to add the following alias:"
  echo "alias ocm-container-stg=\"OCM_URL=staging ocm-container\""
fi

echo
echo
echo "ocm-container configuration can be customized by editing ${CONFIG_DIR}/env.source"

if [[ $( grep -c '^# REQUIRED:' "${CONFIG_DIR}/env.source") -ne 0 ]]; then
  echo
  echo
  echo "it seems that in '${CONFIG_DIR}/env.source' there are some configurations that are not fufilled"
  echo "please remove the REQUIRED line once they are set:"
  echo
  AWK=$( cat << EOF
/^# REQUIRED:/
  {
    print FILENAME ":" NR, \$0;
    tmpfs=FS;
    FS="=";
    getline;
    print FILENAME ":" NR, \$1 "=";
    FS=tmpfs;
    print "";
  }
EOF
)
  awk -f <( echo "$AWK" ) "${CONFIG_DIR}/env.source"
fi
