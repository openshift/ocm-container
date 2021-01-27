#!/bin/bash

source /usr/local/kube_ps1/kube-ps1.sh

## Set Defaults
export EDITOR=vim
export OCM_URL=${OCM_URL:-production}
export PS1="[\W {\[$(tput setaf 2)\]${OCM_URL}\[$(tput sgr0)\]} \$(kube_ps1)]\$ "
export KUBE_PS1_BINARY=oc
export KUBE_PS1_CLUSTER_FUNCTION=cluster_function
export KUBE_PS1_SYMBOL_ENABLE=false
## Overwrite defaults with user-config
source /root/.config/ocm-container/env.source

# make vi work as vim does 
alias vi=vim

complete -C '/usr/local/aws/aws/dist/aws_completer' aws

if [ -n "$INITIAL_CLUSTER_LOGIN" ]
then
  sre-login $INITIAL_CLUSTER_LOGIN
fi

function cluster_function() {
  oc config view  --minify --output 'jsonpath={..server}' | cut -d. -f2-4
}
