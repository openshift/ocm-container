#!/bin/bash

if [ ! -f ./.git/config ]; then
    echo "Not in respository root";
    exit 1;
fi

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
fi

time sudo docker build --no-cache \
  --build-arg osv4client=${osv4client} \
  -t ocm-container .

date
date -u
