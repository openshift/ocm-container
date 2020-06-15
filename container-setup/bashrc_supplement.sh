#!/bin/bash

source /usr/local/kube_ps1/kube-ps1.sh
export PS1='[\u@\h \W $(kube_ps1)]\$ '
export KUBE_PS1_BINARY=oc
export KUBE_PS1_CLUSTER_FUNCTION=cluster_function
export KUBE_PS1_SYMBOL_ENABLE=false

function cluster_function() {
  oc config view  --minify --output 'jsonpath={..server}' | cut -d. -f2-4
}
