#!/usr/bin/env bash

usage() {
  cat <<EOF
  usage: $0 [ OPTIONS ] [ -- Additional Docker Build Options ]
  Options
  -h  --help      Show this message and exit
  -n  --no-cache  Do not use the container runtime cache for images
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
    -t | --tag )            shift
                            BUILD_TAG=$1
                            ;;
    -x | --debug )          set -x
                            ;;

    -- ) shift
      CONTAINER_ARGS+=("$@")
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

# shellcheck disable=SC2006
echo "Building with `build.sh` is deprecated and will be removed in a later version.  Please use `ocm-container build`, or `make build` directly, instead."

### cd locally
cd "$(dirname $0)" || exit

./ocm-container build --tag="${BUILD_TAG}" --build-args="${NOCACHE} ${CONTAINER_ARGS[*]}"