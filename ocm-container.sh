#!/usr/bin/env bash

usage() {
  cat <<EOF
  usage: $0 [ OPTIONS ] [ Initial Cluster Name ]
  Options
  -e  --exec          		Path (in-container) to a script to run on-cluster and exit
  -d --disable-console-port   	Disable automatic cluster console port mapping (Linux only; console port will not work with MacOS)
  -h  --help          		Show this message and exit
  -n  --no-personalizations     Disables personalizations file, typically used for debugging potential issues or during automated script running
  -o  --launch-opts   		Sets extra non-default container launch options
  -t  --tag           		Sets the image tag to use
  -x  --debug         		Set bash debug flag

  Initial Cluster Name can be either the cluster name, cluster id, or external cluster ID.
EOF
}

ARGS=()
BUILD_TAG="latest"
EXEC_SCRIPT=
TTY="-it"
ENABLE_PERSONALIZATION_MOUNT=true

DEFAULT_BACKPLANE_CONFIG_DIR_LOCATION="$HOME/.config/backplane"

while [ "$1" != "" ]; do
  case $1 in
    -e | --exec )                   shift
                                    EXEC_SCRIPT=$1
                                    ;;
    -d | --disable-console-port )   DISABLE_CONSOLE_PORT_MAP=true
                                    ;;
    -h | --help )                   usage
                                    exit 1
                                    ;;
    -n | --no-personalizations )    ENABLE_PERSONALIZATION_MOUNT=false
                                    ;;
    -o | --launch-opts )            shift
                                    if [[ -n "${OCM_CONTAINER_LAUNCH_OPTS}" ]]; then
                                      OCM_CONTAINER_LAUNCH_OPTS="$OCM_CONTAINER_LAUNCH_OPTS $1"
                                    else
                                      OCM_CONTAINER_LAUNCH_OPTS=$1
                                    fi
                                    ;;
    -t | --tag )                    shift
                                    BUILD_TAG="$1"
                                    ;;
    -x | --debug )                  set -x
                                    ;;

    -* ) echo "Unexpected parameter $1"
         usage
         exit 1
         ;;

    * )
      ARGS+=($1)
      ;;
  esac
  shift
done

if [ ${#ARGS[@]} -gt 1 ]
then
  echo "Expected at most one argument.  Got ${#ARGS[@]}"
  usage
  exit 1
fi

### Load config
CONFIG_DIR=${HOME}/.config/ocm-container
export OCM_CONTAINER_CONFIGFILE="$CONFIG_DIR/env.source"

if [ ! -f ${OCM_CONTAINER_CONFIGFILE} ]; then
    echo "Cannot find config file at $OCM_CONTAINER_CONFIGFILE";
    echo "Run the init.sh file to create one."
    echo "exiting"
    exit 1;
fi

source ${OCM_CONTAINER_CONFIGFILE}

### SSH Agent Mounting
operating_system=`uname`

SSH_AGENT_MOUNT="-v ${SSH_AUTH_SOCK}:/tmp/ssh.sock:ro"
SSH_AUTH_SOCK_ENV="-e \"SSH_AUTH_SOCK=/tmp/ssh.sock\""

### Mount ssh sockets dir used for ssh connection multiplexing
SSH_SOCKETS_DIR=${HOME}/.ssh/sockets
if [ -d "${SSH_SOCKETS_DIR}" ] && [ "${DISABLE_SSH_MULTIPLEXING}" != "true" ]
then
 SSH_SOCKETS_MOUNT="-v ${SSH_SOCKETS_DIR}:/root/.ssh/sockets"
fi

if [[ "$CONTAINER_SUBSYS" != "podman" ]] && [[  "$operating_system" == "Darwin" ]]
then
  SSH_AGENT_MOUNT="--mount type=bind,src=/run/host-services/ssh-auth.sock,target=/tmp/ssh.sock,readonly"
elif [[ "$CONTAINER_SUBSYS" == podman ]] && [[ "$operating_system" == "Darwin" ]]
then
  agent_location=$(podman machine ssh 'ls /private/tmp | grep com.apple.launchd')
  SSH_AGENT_MOUNT="-v /private/tmp/$agent_location:/tmp/ssh:ro"
  SSH_AUTH_SOCK_ENV="-e \"SSH_AUTH_SOCK=/tmp/ssh/Listeners\""
  SSH_SOCKETS_MOUNT="--mount type=tmpfs,destination=/root/.ssh/sockets"
fi


### AWS token pull
if [[ -f "${HOME}/.aws/credentials" ]]
then
  AWSFILEMOUNT="-v ${HOME}/.aws/credentials:/root/.aws/credentials:ro"
fi

if [[ -f "${HOME}/.aws/config" ]]
then
  AWSFILEMOUNT="${AWSFILEMOUNT:-''} -v ${HOME}/.aws/config:/root/.aws/config:ro"
fi

### JIRA Token Mounting
JIRA_CONFIG_DIR=".config/.jira/"
if [ -d ${HOME}/${JIRA_CONFIG_DIR} ]
then
  JIRAFILEMOUNT="-v ${HOME}/${JIRA_CONFIG_DIR}:/root/${JIRA_CONFIG_DIR}:ro"
fi

if [ -f ${HOME}/${JIRA_CONFIG_DIR}/token.json ]
then
  JIRATOKENCONFIG="-e JIRA_API_TOKEN=$(jq -r .token ${HOME}/${JIRA_CONFIG_DIR}/token.json) -e JIRA_AUTH_TYPE=bearer"
fi

### PagerDuty Token Mounting
PAGERDUTY_TOKEN_FILE=".config/pagerduty-cli/config.json"
if [ -f ${HOME}/${PAGERDUTY_TOKEN_FILE} ]
then
  PAGERDUTYFILEMOUNT="-v ${HOME}/${PAGERDUTY_TOKEN_FILE}:/root/${PAGERDUTY_TOKEN_FILE}:ro"
fi

### osdctl config mounting
OSDCTL_CONFIG=".config/osdctl"
if [ -f ${HOME}/${OSDCTL_CONFIG} ]
then
  OSDCTL_CONFIG_MOUNT="-v ${HOME}/${OSDCTL_CONFIG}:/root/${OSDCTL_CONFIG}:ro"
fi

### Google Cloud CLI config mounting
if [ -d ${HOME}/.config/gcloud ]; then
  GOOGLECLOUDFILEMOUNT="
-v ${HOME}/.config/gcloud/active_config:/root/.config/gcloud/active_config_readonly:ro
-v ${HOME}/.config/gcloud/configurations/config_default:/root/.config/gcloud/configurations/config_default_readonly:ro
-v ${HOME}/.config/gcloud/credentials.db:/root/.config/gcloud/credentials_readonly.db:ro
-v ${HOME}/.config/gcloud/access_tokens.db:/root/.config/gcloud/access_tokens_readonly.db:ro
"
fi

### OPS-UTILS-DIR File Mounting
if [ -d "${OPS_UTILS_DIR}" ]; then
  OPS_UTILS_DIR_RW_FLAG="ro"
  if [[ ${OPS_UTILS_DIR_RW} = true ]]; then
    OPS_UTILS_DIR_RW_FLAG="rw"
  fi
  OPS_UTILS_DIR_MOUNT="
-v ${OPS_UTILS_DIR}:/root/sop-utils:${OPS_UTILS_DIR_RW_FLAG}
"
fi

### mount a scratch dir
if [ -n "$SCRATCH_DIR" ]
then
  SCRATCH_DIR_MOUNT="-v ${SCRATCH_DIR}:/root/scratch/"
fi

### Automatic Login Detection
if [ -n "$ARGS" ]
then
  INITIAL_CLUSTER_LOGIN="-e INITIAL_CLUSTER_LOGIN=$ARGS"
  ### per-cluster persistent bash histories
  if [ -n "$PERSISTENT_CLUSTER_HISTORIES" ]
  then
    PER_CLUSTER_ID=$(ocm describe cluster "$ARGS"|awk '/^ID:/{print $2}' 2>/dev/null)
    if [ -n "$PER_CLUSTER_ID" ]
    then
      PER_CLUSTER_PERSISTENT="$HOME/.config/ocm-container/per-cluster-persistent/$PER_CLUSTER_ID"
      mkdir -p "$PER_CLUSTER_PERSISTENT"
      OCM_CONTAINER_LAUNCH_OPTS+=" -v $PER_CLUSTER_PERSISTENT:/root/per-cluster:rw"
      OCM_CONTAINER_LAUNCH_OPTS+=" -e HISTFILE=/root/per-cluster/.bash_history"
    fi
  fi
fi

if [ -n "$EXEC_SCRIPT" ]
then
  TTY=""
fi

if [ "${DISABLE_CONSOLE_PORT_MAP}" != "true" ]
then
  PORT_MAP_OPTS="--publish-all"
fi

if [[ $ENABLE_PERSONALIZATION_MOUNT == true ]] && [ -n "$PERSONALIZATION_FILE" ]
then
  if [ -f "$PERSONALIZATION_FILE" ]
  then
    PERSONALIZATION_MOUNT="-v ${PERSONALIZATION_FILE}:/root/.config/personalizations.d/personalizations.sh"
  elif [ -d "$PERSONALIZATION_FILE" ]
  then
    PERSONALIZATION_MOUNT="-v ${PERSONALIZATION_FILE}:/root/.config/personalizations.d"
  else
    echo "Personalizations File is not a valid file or directory. Check your config."
    exit 3

## Check for backplane config dir override and then mount the directory if it exists
if [ -z "$BACKPLANE_CONFIG_DIR" ]
then
  BACKPLANE_CONFIG_DIR=$DEFAULT_BACKPLANE_CONFIG_DIR_LOCATION
fi

if [ -d "$BACKPLANE_CONFIG_DIR" ]
then
  BACKPLANE_CONFIG_MOUNT="-v $BACKPLANE_CONFIG_DIR:/root/.config/backplane"
  if [ -z $OCM_URL ] || [ $OCM_URL == "production" ]
  then
    BACKPLANE_CONFIG_MOUNT="$BACKPLANE_CONFIG_MOUNT -e BACKPLANE_CONFIG=/root/.config/backplane/config.json"
  else
    BACKPLANE_CONFIG_MOUNT="$BACKPLANE_CONFIG_MOUNT -e BACKPLANE_CONFIG=/root/.config/backplane/config.$OCM_URL.json"
  fi
fi

### start container
CONTAINER=$(${CONTAINER_SUBSYS} create $TTY --rm --privileged \
-e "OCM_URL" \
-e "USER" \
-e "OFFLINE_ACCESS_TOKEN" \
${JIRATOKENCONFIG} \
${INITIAL_CLUSTER_LOGIN} \
-v ${CONFIG_DIR}:/root/.config/ocm-container:ro \
-v ${HOME}/.ssh:/root/.ssh:ro \
${GOOGLECLOUDFILEMOUNT} \
${JIRAFILEMOUNT} \
${PAGERDUTYFILEMOUNT} \
${OSDCTL_CONFIG_MOUNT} \
${AWSFILEMOUNT} \
${SSH_AGENT_MOUNT} \
${SSH_SOCKETS_MOUNT} \
${SSH_AUTH_SOCK_ENV} \
${OPS_UTILS_DIR_MOUNT} \
${SCRATCH_DIR_MOUNT} \
${PORT_MAP_OPTS} \
${OCM_CONTAINER_LAUNCH_OPTS} \
<<<<<<< HEAD
${PERSONALIZATION_MOUNT} \
=======
${BACKPLANE_CONFIG_MOUNT} \
>>>>>>> da226e4 (Adds the new public backplane repo)
ocm-container:${BUILD_TAG} ${EXEC_SCRIPT})

$CONTAINER_SUBSYS start $CONTAINER > /dev/null

if [ "${DISABLE_CONSOLE_PORT_MAP}" != "true" ]
then
  TMPDIR=$(mktemp -d)
  echo $($CONTAINER_SUBSYS inspect $CONTAINER \
    | jq -r '.[].NetworkSettings.Ports |select(."9999/tcp" != null) | ."9999/tcp"[].HostPort') > ${TMPDIR}/portmap
  $CONTAINER_SUBSYS cp ${TMPDIR}/portmap $CONTAINER:/tmp/portmap
fi

$CONTAINER_SUBSYS attach $CONTAINER
