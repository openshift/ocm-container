#!/bin/bash

if [ ! -f ./ocm-container.sh ]; then
    echo "Not in source root, cd into correct directory";
    exit 1;
fi

source env.source

sudo docker run -it --rm \
-e "OFFLINE_ACCESS_TOKEN=${OFFLINE_ACCESS_TOKEN}" \
-e "OCM_USER=${OCM_USER}" \
ocm-container /bin/bash ## -c "/container-setup/login.sh $@ && /container-setup/bash-ps1-wrap.sh"
