#!/usr/bin/env bash

#set -eEuo pipefail

# pdsr will show resolved incidents for a service
function pdsr() {
  local service=$1
  pd incident:list -s resolved -S $service
}

SILENT_TEST_USER_ID=P8QS6CC
LONG_RUNNING_INCIDENT_ID=PJWIXM0 
# pdstst will silent-test the incident id provided partially (still need to merge)
function pdstst() {
  local INCIDENTID=$1
  pd incident:assign -u $SILENT_TEST_USER_ID -i $INCIDENTID
  echo "need functionality to merge the incident to the $LONG_RUNNING_INCIDENT_ID, so far just opening it manually and merging via UI"
  pd incident:open -i $LONG_RUNNING_INCIDENT_ID # long term incident to merge into
}

ESCALATION_POLICY_ID=PA4586M
# pdan will assign all pd alerts to the next in the escalation policy
alias pdan='pd incident:list --me -p | pd incident:assign -e ${ESCALATION_POLICY_ID} -p'

# pdis will show raw data for pagerduty incidents, which are relevant to the oncall personnel
function pdis() {
   local I=$1
  pd rest:get -e=/incidents/$I/alerts | jq -r '.alerts[].body.details|., .firing, "notes = \(.notes)", "cluster_id = \(.cluster_id)"'
}

# pdisy will do something similar to pdis but in yaml format
function pdisy() {
  local INCIDENTID=$1
  pd incident:alerts -i $INCIDENTID -j | yq e -P -
}

# pdil will list incidents sorted by param, incidents will owned by the user
function pdil() {
  local filter='Urgency'
  if [[ $# -ne 0 ]]; then
    filter="$1"	
  fi
  sort_by_filter="--sort $filter"

  pd incident:list --me $sort_by_filter 
}

# pdilu will list incidents based on the user
function pdilu() {
  if [[ $# -eq 0 ]]; then
    echo "cannot list pd incidents by user where no user is provided"
    return
  fi
  user="$@"
  
  pd incident:list -e $user 
}

# pdepo will show all of the oncall personnel in the escalation 
ESCALATION_POLICY_ID=PA4586M 
alias pdepo='pd ep:oncall -i ${ESCALATION_POLICY_ID} --sort Level'


IGNORED_USER='Silent Test'
TEAM='Platform SRE'
# pdilt lists incidents that are set to the team and not to the ignored user
alias pdilt='pd incident:list --teams $TEAM--filter "-Assigned to=$IGNORED_USER"'
