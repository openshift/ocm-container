#!/usr/bin/env bash

set -x

usage() {
  cat <<EOF
  usage: $0 [ OPTIONS ] [ -- Additional Docker Build Options ]
  Options
  -h  --help      Show this message and exit
  -n  --no-cache  Do not use the container runtime cache for images
  -p  --platform  Platform to build (ex. linux/amd64; linux/arm64)
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
    -n | --no-cache )       NOCACHE="--no-cache "
                            ;;
    -p | --platform )       shift
                            PLATFORM="$1"
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

OCM_CONTAINER_CONFIG_PATH="${HOME}/.config/ocm-container"
export OCM_CONTAINER_CONFIG="${OCM_CONTAINER_CONFIG_PATH}/env.source"

### Create default config from sample
if [ ! -f ${OCM_CONTAINER_CONFIG} ]; then
    echo "Cannot find config file, creating one from sample...";
    mkdir -pv ${OCM_CONTAINER_CONFIG_PATH}
    cp -v ./env.source.sample ${OCM_CONTAINER_CONFIG}
fi

### Load config
source ${OCM_CONTAINER_CONFIG}

### start build
echo "Using ${CONTAINER_SUBSYS} to build the container"

PLATFORM_BUILD_ARG="build"
### Start multi-platform build handling
if [[ -n $PLATFORM ]] 
then
  #arch="$(echo $PLATFORM | cut -d/ -f2)"
  #system_arch=""
  #case "$(uname -m)" in
  #  aarch64) system_arch="arm64";;
  #  armv8b) system_arch="arm64";;
  #  armv8l) system_arch="arm64";;
  #  arm64) system_arch="arm64";;
  #  amd64) system_arch="amd64";;
  #  x86_64) system_arch="amd64";;
  #  * ) echo "Unexpected system architecture: $(uname -m); exiting"; exit 228;;
  #esac

  #if [[ "$arch" != "$system_arch" ]] && [[ "${CONTAINER_SUBSYS}" == "docker" ]]
  if [[ "$CONTAINER_SUBSYS" == "docker" ]]
  then
    PLATFORM_BUILD_ARG="buildx build --platform=$PLATFORM"
  else
    PLATFORM_BUILD_ARG="build --platform=$PLATFORM"
  fi
fi

# for time tracking
date
date -u

# we want the $@ args here to be re-split
time ${CONTAINER_SUBSYS} $PLATFORM_BUILD_ARG $NOCACHE\
  $CONTAINER_ARGS \
  -t ocm-container:${BUILD_TAG} .

# for time tracking
date
date -u
