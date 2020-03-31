# ocm-container

A quick environment for accessing OpenShift v4 clusters. Nothing fancy, gets the job done.

Related tools added to image:
* `ocm`
* `oc`
* `aws`

Features:
* Uses ephemeral containers per cluster login, keeping seperate `.kube` configuration and credentials.
* Credentials are destroyed on container exit (container has `--rm` flag set)
* Displays current cluster-name in bash PS1

## Usage:

Build:

* export osv4client=openshift-client-linux-4.3.5.tar.gz
* ./build.sh

Config:

* cp env.source.sample env.source
* vim env.source
** set your requested OCM_USER (for `ocm -u OCM_USER`)
** set your OFFLINE_ACCESS_TOKEN (from [cloud.redhat.com](https://cloud.redhat.com/))

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

