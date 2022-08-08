#!/usr/bin/env bash

usage() {
  cat <<EOF
  usage: $0 [ OPTIONS ] [ Initial Cluster Name ]
  Options
  -e  --exec          		Path (in-container) to a script to run on-cluster and exit
  -h  --help          		Show this message and exit
  -o  --launch-opts   		Sets extra non-default container launch options
  -t  --tag           		Sets the image tag to use
  -x  --debug         		Set bash debug flag
  -d --disable-console-port   	Disable automatic cluster console port mapping (Linux only; console port will not work with MacOS)

  Initial Cluster Name can be either the cluster name, cluster id, or external cluster ID.
EOF
}

ARGS=()
BUILD_TAG="latest"
EXEC_SCRIPT=
TTY="-it"

while [ "$1" != "" ]; do
  case $1 in
    -e | --exec )           shift
                            EXEC_SCRIPT=$1
                            ;;
    -h | --help )           usage
                            exit 1
                            ;;
    -o | --launch-opts )    shift
                            if [[ -n "${OCM_CONTAINER_LAUNCH_OPTS}" ]]; then
                              OCM_CONTAINER_LAUNCH_OPTS="$OCM_CONTAINER_LAUNCH_OPTS $1"
                            else
                              OCM_CONTAINER_LAUNCH_OPTS=$1
                            fi
                            ;;
    -t | --tag )            shift
                            BUILD_TAG="$1"
                            ;;
    -x | --debug )          set -x
                            ;;
    -d | --disable-console-port )   DISABLE_CONSOLE_PORT_MAP=true
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

if [[ "$operating_system" == "Darwin" ]]
then
  SSH_AGENT_MOUNT="--mount type=bind,src=/run/host-services/ssh-auth.sock,target=/tmp/ssh.sock,readonly"
fi

### AWS token pull
if [[ -d "${HOME}/.aws" ]]; then
  AWSFILEMOUNT="
-v ${HOME}/.aws/credentials:/root/.aws/credentials:ro
-v ${HOME}/.aws/config:/root/.aws/config:ro
"
fi

### PagerDuty Token Mounting
PAGERDUTY_TOKEN_FILE=".config/pagerduty-cli/config.json"
if [ -f ${HOME}/${PAGERDUTY_TOKEN_FILE} ]
then
  PAGERDUTYFILEMOUNT="-v ${HOME}/${PAGERDUTY_TOKEN_FILE}:/root/${PAGERDUTY_TOKEN_FILE}:ro"
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
fi

if [ -n "$EXEC_SCRIPT" ]
then
  TTY=""
fi

if [ "${DISABLE_CONSOLE_PORT_MAP}" != "true" ]
then
  PORT_MAP_OPTS="--publish-all"
fi

if [ "${DISABLE_SOCKET_MOUNT}" == "true" ]
then
  SOCKET_MOUNT="-v /root/.ssh/sockets"
else
  SOCKET_MOUNT="-v ${HOME}/.ssh/sockets:/root/.ssh/sockets"
fi

### start container
CONTAINER=$(${CONTAINER_SUBSYS} create $TTY --rm --privileged \
-e "OCM_URL" \
-e "USER" \
-e "SSH_AUTH_SOCK=/tmp/ssh.sock" \
-e "OFFLINE_ACCESS_TOKEN" \
${INITIAL_CLUSTER_LOGIN} \
-v ${CONFIG_DIR}:/root/.config/ocm-container:ro \
-v ${HOME}/.ssh:/root/.ssh:ro \
${GOOGLECLOUDFILEMOUNT} \
${PAGERDUTYFILEMOUNT} \
${AWSFILEMOUNT} \
${SSH_AGENT_MOUNT} \
${OPS_UTILS_DIR_MOUNT} \
${SCRATCH_DIR_MOUNT} \
${SOCKET_MOUNT} \
${PORT_MAP_OPTS} \
${OCM_CONTAINER_LAUNCH_OPTS} \
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
