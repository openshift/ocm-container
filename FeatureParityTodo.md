# Feature Parity TODO

These must be implemented in the new ocm-container golang CLI to meet acceptance criteria

## Flags:

* -e, --exec ✅
* -d, --disable-console-port
* -h, --help ✅
* -n, --no-personalizations ✅
* -o, --launch-opts
* -t, --tag ✅
* -x, --debug ✅

## SSH

Re-evaluate if the SSH_AGENT_MOUNT stuff is necessary

* Check for Global ENV for SSH_AUTH_SOCK - https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L104
* SSH Sock volume mount for agent - https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L104-L110
* SSH_AUTH_SOCK forr multiplexing - https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L114-L118
* Podman machine and Mac support - https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L121-130
* Mount homedir .ssh -  https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L298

## CA Anchors

Volume mount for CAs:

https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L114-L118

## AWS

Volume Mounts: 

* AWS Credentials
* AWS Config

https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L133-L142

## Jira

Volume Mount: 
* Jira Config Dir

Env Variable: 

* Jira token

https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L144-L156

## PagerDuty Token

Volume Mount
https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L160-L164

## OSDCTL

Volume Mount
https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L167-L171

## GCloud

Volume Mount
https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L174-L177

## OPS SOP

Volume Mount

* Read only
* Read write
* Check Global ENV for OPS_UTILS_DIR_RW_FLAG
* Evaluate if this is the right way to do this

https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L180-L188

## Scratch Dir

Volume mount, Check Global ENV
https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L191-L194

## Persistent cluster histories

* Check for Global ENV PERSISTENT_CLUSTER_HISTORIES
* Check OCM Output for cluster id (exec an OCM command)
* Make a HOMEDIR directory
* Volume mount
* Env for HISTFILE

https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L197-L221


## Exec Script

* Exec the script, disable TTY

## Console

* publish ports
* copy file inside container after launch

## Personalization

* check for env; Check for file
* mount script or directory

https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L224-L235

## Backplane

* Config dir
* Set OCM URL
* Mount Config Dir
* Set config.json
* CONFIG - setup backplane config (this can be part of the config and not the init)

## Generic 

* USER env
* OCM_URL env
* OFFLINE_ACCESS_TOKEN env
