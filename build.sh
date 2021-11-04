#!/usr/bin/env bash

usage() {
  cat <<EOF
  usage: $0 [ OPTIONS ] [ -- Additional Docker Build Options ]
  Options
  -h  --help      Show this message and exit
  -t  --tag       Build with a specific docker tag
  -x  --debug     Set the bash debug flag

  Example:

  $0 --tag devel -- "--build-arg=OSDCTL_VERSION=tags/v0.4.0 --build-arg=ROSA_VERSION=v1.0"

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

### start build

# for time tracking
date
date -u

# we want the $@ args here to be re-split
time ${CONTAINER_SUBSYS} build \
  $CONTAINER_ARGS \
  -t ocm-container:${BUILD_TAG} .

# for time tracking
date
date -u
