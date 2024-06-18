#!/usr/bin/env bash
source ${HOME}/.bashrc.d/05-kube-ps1.sh

function cluster_function() {
  info="$(ocm backplane status 2> /dev/null)"
  if [ $? -ne 0 ]; then return; fi
  clustername=$(grep "Cluster Name" <<< $info | awk '{print $3}')
  baseid=$(grep "Cluster Basedomain" <<< $info | awk '{print $3}' | cut -d'.' -f1,2)
  echo $clustername.$baseid
}

export PS1="[\W {\[\033[1;32m\]${OCMC_OCM_URL}\[\033[0m\]} \$(kube_ps1)]\$ "
export KUBE_PS1_BINARY=oc
export KUBE_PS1_CLUSTER_FUNCTION=cluster_function
export KUBE_PS1_SYMBOL_ENABLE=false
