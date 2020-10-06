#!/bin/bash -e

if [ "$I_AM_IN_CONTAINER" != "I-am-in-container" ]; then
  echo "must be run in container";
  exit 1;
fi

echo "in container";

#export osv4client=openshift-client-linux-4.3.5.tar.gz

# https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/
mkdir /usr/local/oc;
pushd /usr/local/oc;
wget -q https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/${osv4client};
tar xzvf ${osv4client};
rm ${osv4client};
ln -s /usr/local/oc/oc /usr/local/bin/oc;
oc completion bash >  /etc/bash_completion.d/oc
popd;
