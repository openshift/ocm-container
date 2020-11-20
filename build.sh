#!/bin/bash

usage() {
  cat <<EOF
  usage: $0 [ OPTIONS ] [ -- Additional Docker Options ]
  Options
  -h  --help      Show this message and exit
  -t  --tag       Build with a specific docker tag
  -x  --debug     Set the bash debug flag
EOF
}

BUILD_TAG="latest"
CONTAINER_ARGS=()

while [ "$1" != "" ]; do
  case $1 in
    -h | --help )           usage
                            exit 1
                            ;;
    -t | --tag )            shift
                            BUILD_TAG=$1
                            ;;
    -x | --debug )          set -x
                            ;;

    -- ) shift
      CONTAINER_ARGS+=($@)
      break
      ;;

    -* ) echo "Unexpected parameter $1"
        usage
        exit 1
        ;;

    * ) echo "Unexpected parameter $1"
        usage
        exit 1
  esac
  shift
done

### cd locally
cd $(dirname $0)

### Load config
export OCM_CONTAINER_CONFIG="${HOME}/.config/ocm-container/env.source"

export CONTAINER_SUBSYS="sudo docker"

if [ ! -f ${OCM_CONTAINER_CONFIG} ]; then
    echo "Cannot find config file, exiting";
    exit 1;
fi

source ${OCM_CONTAINER_CONFIG}

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
fi

### start build

# for time tracking
date
date -u

# we want the $@ args here to be re-split
time ${CONTAINER_SUBSYS}  build \
  --build-arg osv4client=${osv4client} \
  $CONTAINER_ARGS \
  -t ocm-container:${BUILD_TAG} .

# for time tracking
date
date -u
