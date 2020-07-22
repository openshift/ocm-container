#!/bin/bash

### This script is a temporary change until the group is changed
### But it will allow you to preform cluster admin operations
###
### USE WITH CAUTION
oc adm groups add-users osd-sre-cluster-admins $(oc whoami)
