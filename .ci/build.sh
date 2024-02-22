#!/usr/bin/env bash

build_cmds=(
  "bash build.sh -t latest-arm64 -- '--platform=linux/arm64'"
  "bash build.sh -t latest-amd64 -- '--platform=linux/amd64'"
)

if command -v parallel &> /dev/null
then
  (for cmd in "${build_cmds[@]}"
  do
    echo "echo '$cmd' && $cmd && echo && echo '-----'"
  done) | parallel
else
  for cmd in "{build_cmds[@]}"
  do
    eval "$cmd"
    echo
    echo "----"
    echo
  done
fi


