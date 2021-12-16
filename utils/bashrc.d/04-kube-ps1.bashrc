#!/usr/bin/env bash

set -eEo pipefail # remove nounset as is causes issues

source ${HOME}/.bashrc.d/05-kube-ps1.sh

export PS1="[\W {\[\033[1;32m\]${OCM_URL}\[\033[0m\]} \$(kube_ps1)]\$ "
export KUBE_PS1_BINARY=oc
export KUBE_PS1_CLUSTER_FUNCTION=cluster_function
export KUBE_PS1_SYMBOL_ENABLE=false
export KUBE_PS1_ENABLED=on

#set +eEo pipefail 
