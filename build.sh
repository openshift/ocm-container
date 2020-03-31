#!/bin/bash

echo "Check the following URL for latest available OpenShift client:"
echo "https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/"
echo
echo "export osv4client=openshift-client-linux-4.3.5.tar.gz"
echo ${0}
echo
echo

sudo docker build --no-cache \
  --build-arg osv4client=${osv4client} \
  -t ocm-container .
