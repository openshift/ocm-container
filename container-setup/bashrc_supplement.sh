#!/bin/bash

source /usr/local/kube_ps1/kube-ps1.sh
export PS1='[\u@\h \W $(ocm_environment) $(kube_ps1)]\$ '
export KUBE_PS1_BINARY=oc
export KUBE_PS1_CLUSTER_FUNCTION=cluster_function
export KUBE_PS1_SYMBOL_ENABLE=false

function cluster_function() {
  oc config view  --minify --output 'jsonpath={..server}' | cut -d. -f2-4
}

function ocm_environment() {
	# based on how ocm-cli works for now, when the default change we will go with it
	OCM_URL=${OCM_URL:-prdoduction}
	echo "{$(tput setaf 2)${OCM_URL}$(tput sgr0)}"
}
