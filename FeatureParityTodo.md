# Feature Parity TODO

These must be implemented in the new ocm-container golang CLI to meet acceptance criteria

## Flags:

* -e, --exec ✅
* -d, --disable-console-port ✅
* -h, --help ✅
* -n, --no-personalizations ✅
* -o, --launch-opts
* -t, --tag ✅
* -x, --debug ✅

## SSH

Re-evaluate if the SSH_AGENT_MOUNT stuff is necessary

* Check for Global ENV for SSH_AUTH_SOCK - https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L104
* SSH Sock volume mount for agent - https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L104-L110
* SSH_AUTH_SOCK for multiplexing - https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L114-L118
* Podman machine and Mac support - https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L121-130
* Mount homedir .ssh -  https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L298


NOTES: We've removed SSH agent mounting and multiplexing, because while we still may SSH from ocm-container to a jumphost, the 
removal of SSHUTTLE means we're considerably less likely to need any multiplexing, and SSH keys are most likely to be fetched 
from HIVE, rather than anything stored on the local machine, and in fact, it would be more of a risk to mount these.

## CA Anchors DONE ✅ 

Volume mount for CAs:

https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L114-L118

## AWS DONE ✅ 

Volume Mounts: 

* AWS Credentials
* AWS Config

https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L133-L142

## Jira DONE ✅ 

Volume Mount: 
* Jira Config Dir

Env Variable: 

* Jira token

https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L144-L156

## PagerDuty Token DONE ✅ 

Volume Mount
https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L160-L164

## OSDCTL DONE ✅ 

Volume Mount
https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L167-L171

## GCloud DONE ✅ 

Volume Mount
https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L174-L177

## OPS SOP DONE ✅ 

Volume Mount

* Read only
* Read write
* Check Global ENV for OPS_UTILS_DIR_RW_FLAG
* Evaluate if this is the right way to do this

https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L180-L188

## Scratch Dir DONE ✅ 

Volume mount, Check Global ENV
https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L191-L194

## Persistent cluster histories

* Check for Global ENV PERSISTENT_CLUSTER_HISTORIES
* Check OCM Output for cluster id (exec an OCM command)
* Make a HOMEDIR directory
* Volume mount
* Env for HISTFILE

https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L197-L221


NOTES: This feature made use of OCM _outside_ of the container, and so is not easily implementable with a fully containerized ocm-container.  Without being able to lookup the OCM description of the cluster (which we cannot do without a token), we cannot mount the right directory into the container for the per-cluster histories.  

This MIGHT be doable with external tokens, which we'd have access to, and OCM API calls directly within the go code.  More thought is needed for this.

## Exec Script DONE ✅ 

* Exec the script, disable TTY

## Console  DONE ✅ 

* publish ports
* copy file inside container after launch

## Personalization  DONE ✅ 

* check for env; Check for file
* mount script or directory

https://github.com/openshift/ocm-container/blob/dfeac58f52a5a9c4e3bb36ad6cfe3bbdb27c9e12/ocm-container.sh#L224-L235

## Backplane DONE ✅ 

* Config dir
* Set OCM URL
* Mount Config Dir
* Set config.json
* CONFIG - setup backplane config (this can be part of the config and not the init)

## Generic DONE ✅ 

* USER env
* OCM_URL env
* OFFLINE_ACCESS_TOKEN env
