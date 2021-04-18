#!/bin/bash -e

if [ "$I_AM_IN_CONTAINER" != "I-am-in-container" ]; then
  echo "must be run in container";
  exit 1;
fi

echo "in container";

### Select osv4client version, auto-detect from mirror.openshift.com
if [ "x${osv4client}" == "x" ]; then
    # auto-detect latest openshift-client-linux-4.x.y.tar.gz
    osv4clienturl="https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/"
    export osv4client=`curl -s -q ${osv4clienturl} \
        | grep openshift-client-linux-4 | grep .tar.gz  \
        | sed -E 's/.*(openshift-client-linux-4.+tar.gz).*/\1/g'`

    echo "Check the following URL for latest available OpenShift client:"
    echo ${osv4clienturl}
    echo
    echo "using:"
    echo "export osv4client=${osv4client}"
    echo ${0}
    echo
    exit 1
fi

mkdir /usr/local/oc;
pushd /usr/local/oc;
wget -q https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/${osv4client};
tar xzvf ${osv4client};
rm ${osv4client};
ln -s /usr/local/oc/oc /usr/local/bin/oc;
oc completion bash >  /etc/bash_completion.d/oc
popd;
