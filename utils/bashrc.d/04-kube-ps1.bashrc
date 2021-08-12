#!/usr/bin/env bash
source ${HOME}/.bashrc.d/05-kube-ps1.sh

export PS1="[\W {\e[32m${OCM_URL}\e[39m} \$(kube_ps1)]\$ "
export KUBE_PS1_BINARY=oc
export KUBE_PS1_CLUSTER_FUNCTION=cluster_function
export KUBE_PS1_SYMBOL_ENABLE=false
