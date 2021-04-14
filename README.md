- [ocm-container](#ocm-container)
  - [Features:](#features)
  - [Getting Started:](#getting-started)
    - [Config:](#config)
    - [Build:](#build)
    - [Use it:](#use-it)
  - [Example:](#example)
    - [Public Clusters](#public-clusters)
    - [Private clusters](#private-clusters)

# ocm-container

A quick environment for accessing OpenShift v4 clusters. Nothing fancy, gets the job done.

Related tools added to image:
* `ocm`
* `oc`
* `aws`
* `osdctl`

## Features:
* Does not mount any host filesystem objects as read/write, only uses read-only mounts.
* Uses ephemeral containers per cluster login, keeping seperate `.kube` configuration and credentials.
* Credentials are destroyed on container exit (container has `--rm` flag set)
* Displays current cluster-name, and OpenShift project (`oc project`) in bash PS1
* Ability to login to private clusters without using a browser

## Quick Start:

* clone this repo
* `./init.sh`
* `./build.sh` (while this is building you can run the rest in another tab until the final command)
* edit the file `$HOME/.config/ocm-container/env.source`
  * set your requested OCM_USER (for `ocm -u OCM_USER`)
  * set your OFFLINE_ACCESS_TOKEN (from [cloud.redhat.com](https://cloud.redhat.com/))
  * set your kerberos username if it's different than your OCM_USER
* optional: configure alias in `~/.bashrc`
  * alias ocm-container-stg="OCM_URL=stg ocm-container"
  * alias ocm-container-local='OCM_CONTAINER_LAUNCH_OPTS="-v $(pwd):/root/local" ocm-container'
* MacOS users: Ensure your kerberos ticket gets sent to a file and not the keychain
  * in your `~/.bashrc`: `export KRB5CCNAME=/tmp/krb5cc` or wherever you want your kerberos ticket to live
* Connect to the VPN
* `kinit -V`
  * MacOS users: here you should see it say `FILE:/tmp/krb5cc` and not `API:{some uuid}`.  If it says `API` follow the troubleshooting steps below
* `ocm-container {cluster-name}`

## Build:

```
./build.sh
```

Build accepts the following flags:
```
  -h  --help      Show this message and exit
  -t  --tag       Build with a specific docker tag
  -x  --debug     Set the bash debug flag
```

You can also override the container build flags by separating them at the end of the command with `--`.  Example:
```
./build.sh -t local-dev -- --no-cache
```

## Use it:
```
ocm-container
```
With launch options:
```
OCM_CONTAINER_LAUNCH_OPTS="-v ~/work/myproject:/root/myproject" ocm-container
--
or
--
ocm-container -o "-v ~/work/myproject:/root/myproject"
```

Launch options provide you a way to add other volumes, add environment variables, or anything else you would need to do to run ocm-container the way you want to. 

_NOTE_: Using the flag for launch options will then NOT use the environment variable `OCM_CONTAINER_LAUNCH_OPTS`

## Automatic Login to a cluster:
```
ocm-container my-cluster-id
```

## Example:

### Public Clusters

```
$ ocm-container
[production] # ./login.sh
[production] # ocm cluster login test-cluster
Will login to cluster:
 Name: test-cluster
 ID: 01234567890123456789012345678901
Authentication required for https://api.test-cluster.shard.hive.example.com:6443 (openshift)
Username: my_user
Password:
Login successful.

You have access to 67 projects, the list has been suppressed. You can list all projects with 'oc projects'

Using project "default".
Welcome! See 'oc help' to get started.
[production] (test-cluster) #
```

### Private clusters
This tool also can tunnel into private clusters.

```
$ ocm-container-stg
[staging] # ./login.sh
[staging] # ocm tunnel --cluster test-cluster -- --dns &
Will create tunnel to cluster:
 Name: test-cluster
 ID: 01234567890123456789012345678901

# /usr/bin/sshuttle --remote sre-user@ssh-url.test-cluster.mycluster.com 10.0.0.0/16 172.31.0.0/16 --dns
client: Connected.
[staging] # cluster-login -c 01234567890123456789012345678901
Login successful.

You have access to 67 projects, the list has been suppressed. You can list all projects with 'oc projects'

Using project "default".
Welcome! See 'oc help' to get started.
[staging] (test-cluster) #
```

### Automatic Login
We've built in functionality to simplify the cluster login steps.  Now within the contianer you can run `sre-login cluster-id` and it will refresh your ocm login, create a tunnel within the container if necessary, and then log-in to the cluster.

`sre-login` accepts both a cluster-name or a cluster-id.  If the cluster-name is not unique, it will not ask which one, but display the clusters and exit.

### Advanced scripting with ocm-container
We've recently added the ability to run a script within the container so that you can run ocm-container within a script.

Given the following shell script saved on the local machine in `~/myproject/in-container.sh`:
```
cat ~/myproject/in-container.sh
#!/bin/bash

# source this so we get all of the goodness of ocm-container
source /root/.bashrc

# get the version of the cluster
oc version >> report.txt
```

We can run that on-container with the following script which runs on the host (~/myproject/on-host.sh):
```
cat ~/myproject/on-host.sh
#!/bin/bash

while read -r cluster_id
do
    echo "Cluster $cluster_id Information:" >> report.txt
    ocm-container -o "-v ${HOME}/myproject/script.sh:/root/script.sh -v ${HOME}/myproject/report.txt:/root/report.txt" -e /root/script.sh $cluster_id
    echo "----"
done < clusters.txt
```

Would loop through all clusters listed in `clusters.txt` and then run `oc version` on the cluster, and add the output into report.txt and then it would exit the container, and move to the next container and do the same.

## Troubleshooting

### SSH Config
If you're on a mac and you get an error similar to:
```Cluster is internal. Initializing Tunnel... /root/.ssh/config: line 34: Bad configuration option: usekeychain```
you might need to add something similar to the following to your ssh config:

```
> cat ~/.ssh/config | head
IgnoreUnknown   UseKeychain,AddKeysToAgent
Host *
  <snip>
  UseKeychain yes
```

UseKeychain is a MacOS specific directive which may cause issues on the linux container that ocm-container runs within.  Adding the `IgnoreUnknown UseKeychain` directive tells the ssh config to ignore that directive when it's unknown so it will not throw errors.

### Kerberos Ticket

Using the browserless login feature currently requires the kinit program to generate a kerberos ticket.  I use the following command (outside the container):

```
kinit
```

If you're getting errors like `401 Client Error: Unauthorized` either your Kerberos ticket has expired or if it's your first time running ocm-container then it might not be configured correctly.

On a Mac and on some linux systems we've seen that it doesn't follow the default kinit functionality where /tmp/krb5cc_$UID is the default cache file location, so you have to explicitly set it with an env var.

Add the following to your bashrc: `export KRB5CCNAME=/tmp/krb5cc`

When troubleshooting, if there's already a kerberos ticket issued (even if it's expired) it will continue to put refreshed kerberos tickets in that location.  If you `klist` and see a ticket with a cache location of `API:{UUID}` or `KEYRING:something:something:something`, or basically anything that's not `FILE:/path`, then run `kdestroy -A` to remove all previous cache files, and run `kinit -V` to display where it's outputting the cache file.
