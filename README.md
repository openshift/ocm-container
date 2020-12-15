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

## Getting Started:

### Config:

* clone this repo
* `./init.sh`
* edit the file `$HOME/.config/ocm-container/env.source`
  * set your requested OCM_USER (for `ocm -u OCM_USER`)
  * set your OFFLINE_ACCESS_TOKEN (from [cloud.redhat.com](https://cloud.redhat.com/))
  * set your kerberos username if it's different than your OCM_USER
* optional: configure alias in `~/.bashrc`
  * alias ocm-container-stg="OCM_URL=stg ocm-container"
  * alias ocm-container-local='OCM_CONTAINER_LAUNCH_OPTS="-v $(pwd):/root/local" ocm-container'

### Build:

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

### Use it:
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

### Automatic Login to a cluster:
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

Tunneling to private clusters requires you to run the kinit program to generate a kerberos ticket. (I'm not sure if it needs the -f flag set for forwardability, but I've been setting it).  I use the following command (outside the container):

```
kinit -f -c $KRB5CCFILE
```

where $KRB5CCFILE is exported to `/tmp/krb5cc` in my .bashrc.

You can also set defaults on forwardability or cache file location, however that's outside the scope of `ocm-container`.

On a Mac, it seems that it doesn't follow the default kinit functionality where /tmp/krb5cc_$UID is the default cache file location, so you have to explicitly set it with an env var.  If you're troubleshooting this, it might help to run `kdestroy -A` to remove all previous cache files, and run `kinit` with the `-V` to display where it's outputting the cache file.  On my machine, it was originally attempting to put this into an API location that's supposed to be windows specific.

### Automatic Login
We've built in functionality to simplify the cluster login steps.  Now within the contianer you can run `sre-login cluster-id` and it will refresh your ocm login, create a tunnel within the container if necessary, and then log-in to the cluster.

`sre-login` accepts both a cluster-name or a cluster-id.  If the cluster-name is not unique, it will not ask which one, but display the clusters and exit.
