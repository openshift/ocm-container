#!/usr/bin/env bash
source ${HOME}/.bashrc.d/05-kube-ps1.sh

export PS1="[\W {\[$(tput setaf 2)\]${OCM_URL}\[$(tput sgr0)\]} \$(kube_ps1)]\$ "
export KUBE_PS1_BINARY=oc
export KUBE_PS1_CLUSTER_FUNCTION=cluster_function
export KUBE_PS1_SYMBOL_ENABLE=false
