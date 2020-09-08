#!/bin/bash

### cd locally
cd $(dirname $0)

### Load config
export OCM_CONTAINER_CONFIG="./env.source"

export CONTAINER_SUBSYS="sudo docker"

if [ ! -f ${OCM_CONTAINER_CONFIG} ]; then
    echo "Cannot find config file, exiting";
    exit 1;
fi

source ${OCM_CONTAINER_CONFIG}

### start container
${CONTAINER_SUBSYS} run -it --rm --privileged \
-e "OCM_URL=${OCM_URL}" \
-e "SSH_AUTH_SOCK=/tmp/ssh.sock" \
-v $(pwd)/env.source:/root/env.source \
-v ${SSH_AUTH_SOCK}:/tmp/ssh.sock \
-v ${HOME}/.ssh:/root/.ssh \
ocm-container ${SSH_AUTH_ENABLE} /bin/bash ## -c "/container-setup/login.sh $@ && /container-setup/bash-ps1-wrap.sh"

