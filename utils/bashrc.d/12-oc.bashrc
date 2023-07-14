#!/usr/bin/env bash
oc-elevate() {
  # set timeout and messaging
  TIMEOUT="60"
  PROMPT_TEXT="Enter justification for performed elevated action (${TIMEOUT}s timeout):"
  NOOP_TEXT="No justification for elevated action was given exiting..."

  # wait for user input 
  read -t "${TIMEOUT}" -p "${PROMPT_TEXT}" REASON

  # if reason is given run elevated action
  if [[ ${REASON} ]]
  then
    ocm backplane elevate "${REASON[@]}" -- "$@"
  else
    # when reason is not given, show no op message
    echo -e ${NOOP_TEXT} >&2
  fi
}
