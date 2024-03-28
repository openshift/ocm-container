#!/usr/bin/env bash

set -eo pipefail

NOCACHE=""
while [ "$1" != "" ]; do
  case $1 in
    -n | --no-cache )       NOCACHE="--no-cache "
                            ;;
    * ) echo "Unexpected parameter $1"
        usage
        exit 1
  esac
  shift
done

# set this here so that we can allow the argument parsing above
set -u

build_cmds=(
#  "bash build.sh -t latest-arm64 $NOCACHE -- '--platform=linux/arm64'"
  "bash build.sh -t latest-amd64 $NOCACHE -- '--platform=linux/amd64'"
)

if command -v parallel &> /dev/null
then
  echo "Running with GNU Parallel. No output will appear until the subprocess has finished."
  (for cmd in "${build_cmds[@]}"
  do
    echo "echo '$cmd' && $cmd && echo && echo '-----'"
  done) | parallel
else
  for cmd in "${build_cmds[@]}"
  do
    eval "$cmd"
    echo
    echo "----"
    echo
  done
fi
