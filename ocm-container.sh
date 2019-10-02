#!/bin/bash

source env.source

sudo docker run -it --rm \
-e "OFFLINE_ACCESS_TOKEN=${OFFLINE_ACCESS_TOKEN}" \
ocm-container /bin/bash -c "/container-setup/login.sh $@ && PS1='[\u@\h \W] ($1)\\$ ' /bin/bash"
