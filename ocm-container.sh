#!/bin/bash

usage() {
  cat <<EOF
  usage: $0 [ OPTIONS ] [ Initial Cluster Name ]
  Options
  -e  --exec          Path (in-container) to a script to run on-cluster and exit
  -h  --help          Show this message and exit
  -o  --launch-opts   Sets extra non-default container launch options
  -t  --tag           Sets the image tag to use
  -x  --debug         Set bash debug flag

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
                            OCM_CONTAINER_LAUNCH_OPTS=$1
                            ;;
    -t | --tag )            shift
                            BUILD_TAG="$1"
                            ;;
    -x | --debug )          set -x
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

### PagerDuty Token Mounting
PAGERDUTY_TOKEN_FILE=".config/pagerduty-cli/config.json"
if [ -f ${HOME}/${PAGERDUTY_TOKEN_FILE} ]
then
  PAGERDUTYFILEMOUNT="-v ${HOME}/${PAGERDUTY_TOKEN_FILE}:/root/${PAGERDUTY_TOKEN_FILE}:ro"
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


### start container
${CONTAINER_SUBSYS} run $TTY --rm --privileged \
-e "OCM_URL" \
-e "USER" \
-e "SSH_AUTH_SOCK=/tmp/ssh.sock" \
-e "OFFLINE_ACCESS_TOKEN" \
${INITIAL_CLUSTER_LOGIN} \
-v ${CONFIG_DIR}:/root/.config/ocm-container:ro \
-v ${HOME}/.ssh:/root/.ssh:ro \
-v ${HOME}/.aws/credentials:/root/.aws/credentials:ro \
-v ${HOME}/.aws/config:/root/.aws/config:ro \
-v ${HOME}/.config/gcloud/active_config:/root/.config/gcloud/active_config:ro \
-v ${HOME}/.config/gcloud/configurations/config_default:/root/.config/gcloud/configurations/config_default:ro \
-v ${HOME}/.config/gcloud/credentials.db:/root/.config/gcloud/credentials.db:ro \
${PAGERDUTYFILEMOUNT} \
${SSH_AGENT_MOUNT} \
${OCM_CONTAINER_LAUNCH_OPTS} \
ocm-container:${BUILD_TAG} ${EXEC_SCRIPT}
