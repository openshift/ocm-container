# shellcheck shell=bash
#
function cluster_function() {
  info="$(ocm backplane status 2> /dev/null)"
  if [ $? -ne 0 ]; then return; fi
  clustername=$(grep "Cluster Name" <<< $info | awk '{print $3}')
  baseid=$(grep "Cluster Basedomain" <<< $info | awk '{print $3}' | cut -d'.' -f1,2)
  echo $clustername.$baseid
}

export KUBE_PS1_BINARY=oc
export KUBE_PS1_SYMBOL_ENABLE=false
export KUBE_PS1_CLUSTER_FUNCTION=cluster_function
export PS1="[\W {\[\033[1;32m\]\$(ocm config get url)\[\033[0m\]} \$(kube_ps1)]\$ "
