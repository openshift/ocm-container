#!/bin/bash

if [ ! -f ./.git/config ]; then
    echo "Not in respository root";
    exit 1;
fi

if [ "x${osv4client}" == "x" ]; then
    echo "Check the following URL for latest available OpenShift client:"
    echo "https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/"
    echo
    echo "export osv4client=openshift-client-linux-4.3.5.tar.gz"
    echo ${0}
    echo
    exit 1
fi

time sudo docker build --no-cache \
  --build-arg osv4client=${osv4client} \
  -t ocm-container .
