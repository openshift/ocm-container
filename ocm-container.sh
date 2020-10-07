#!/bin/bash

### cd locally
cd $(dirname $0)

### Load config
CONFIG_DIR=${HOME}/.config/ocm-container
export OCM_ENV_CONFIGFILE="$CONFIG_DIR/env.source"

if [ ! -f ${OCM_ENV_CONFIGFILE} ]; then
    echo "Cannot find config file at $OCM_CONTAINER_CONFIG";
    echo "Run the init.sh file to create one."
    echo "exiting"
    exit 1;
fi

source ${OCM_ENV_CONFIGFILE}

### start container
${CONTAINER_SUBSYS} run -it --rm --privileged \
-e "OCM_URL=${OCM_URL}" \
-e "SSH_AUTH_SOCK=/tmp/ssh.sock" \
-v ${CONFIG_DIR}:/root/.config/ocm-container \
-v ${SSH_AUTH_SOCK}:/tmp/ssh.sock \
-v ${HOME}/.ssh:/root/.ssh \
-v ${HOME}/.aws/credentials:/root/.aws/credentials \
ocm-container ${SSH_AUTH_ENABLE} /bin/bash 
