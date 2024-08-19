# shellcheck shell=bash
export PS1="[\W {\[\033[1;32m\]\$(ocm config get url)\[\033[0m\]} \$(kube_ps1)]\$ "
export KUBE_PS1_BINARY=oc
export KUBE_PS1_CLUSTER_FUNCTION=cluster_function
export KUBE_PS1_SYMBOL_ENABLE=false