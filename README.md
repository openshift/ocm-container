# ocm-container

A quick environment for accessing OpenShift v4 clusters. Nothing fancy, gets the job done.

Related tools added to image:
* `ocm`
* `oc`
* `aws`

Features:
* Does not mount any host filesystem objects, read only, or read/write
* Uses ephemeral containers per cluster login, keeping seperate `.kube` configuration and credentials.
* Credentials are destroyed on container exit (container has `--rm` flag set)
* Displays current cluster-name in bash PS1

## Usage:

Build:

* Check current version from here, use filename in `osv4client`:
[mirror.openshift.com](https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/)

```
export osv4client=openshift-client-linux-4.3.8.tar.gz
./build.sh
```

Config:

* cp env.source.sample env.source
* vim env.source
  * set your requested OCM_USER (for `ocm -u OCM_USER`)
  * set your OFFLINE_ACCESS_TOKEN (from [cloud.redhat.com](https://cloud.redhat.com/))
* optional: configure alias in `~/.bashrc`
  * alias ocm-container="/path/to/ocm-container/ocm-container.sh"

Usage:

```
# Get a list of clusters.
./ocm-container.sh

# Login to cluster
./ocm-container.sh <cluster_name>

# Multiple logins to multiple clusters
Session 1:
./ocm-container.sh <cluster_name1>

Session 2:
./ocm-container.sh <cluster_name2>

Sesison 3:
./ocm-container.sh <cluster_name3>
```

Example:

```
./ocm-container.sh test-cluster
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
[root@999999999999 /] (test-cluster)#
```
