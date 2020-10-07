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

## Getting Started:

### Config:

* clone this repo
* `./init.sh`
* edit the file `$HOME/.config/ocm-container/env.source`
  * set your requested OCM_USER (for `ocm -u OCM_USER`)
  * set your OFFLINE_ACCESS_TOKEN (from [cloud.redhat.com](https://cloud.redhat.com/))
* optional: configure alias in `~/.bashrc`
  * alias ocm-container-stg="OCM_URL=staging ocm-container"

### Build:

```
./build.sh
```

### Use it:
```
ocm-container
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
[staging] # ocm tunnel --cluster test-cluster -- --dns
Will create tunnel to cluster:
 Name: test-cluster
 ID: 01234567890123456789012345678901

# /usr/bin/sshuttle --remote sre-user@ssh-url.test-cluster.mycluster.com 10.0.0.0/16 172.31.0.0/16 --dns
client: Connected.
^Z
[1]+  Stopped     ocm tunnel --cluster test-cluster -- --dns
[staging] # bg 1
[1]+ ocm tunnel --cluster test-cluster -- --dns &
[staging] # ocm cluster login test-cluster --console
Will login to cluster:
 Name: test-cluster
 ID: 01234567890123456789012345678901
 Console URL: https://console.apps.test-cluster.mycluster.com
[staging] # oc login --token AAABBBCCCDDDEEEFFFGGGHHH myuser@api.apps.test-cluster.mycluster.com
Login successful.

You have access to 67 projects, the list has been suppressed. You can list all projects with 'oc projects'

Using project "default".
Welcome! See 'oc help' to get started.
[staging] (test-cluster) #
```

with the clusterID, you only need to:
- copy the `sshuttle` command to another terminal
- grab the token
- run the ocm cluster login command again and use it to log in
