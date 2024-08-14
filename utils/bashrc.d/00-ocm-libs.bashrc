# shellcheck shell=bash

## NOTE: This file is intended for functions/aliases, etc that are
## NOT executed automatically on login.

# Wrap the ocm backplane console command to handle automation for
# port mapping inside the container
ocm() {
  if [[ "${*}" =~ "backplane console" ]]
  then
    shift 2
    echo "/root/.local/bin/cluster-console ${*}"
    command /root/.local/bin/cluster-console
  else
    command ocm "$@"
  fi
}
