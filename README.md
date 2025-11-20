# ocm-container

An standardized environment for accessing OpenShift v4 clusters.

## Features

* Uses ephemeral containers per cluster login, keeping `.kube` configuration and credentials separate.
* Credentials are destroyed on container exit (container has `--rm` flag set)
* Displays current cluster-name, and OpenShift project (`oc project`) in bash PS1
* Infinitely extendable!
    * Add your own volume mounts or environment variables via command line flags or configuration file for persistence.
    * Create your own Containerfile and reference `FROM: ocm-container:latest` and add whatever binaries you want on top

## Installation

First, download the latest release for your OS/Architecture: [https://github.com/openshift/ocm-container/releases](https://github.com/openshift/ocm-container/releases)

This is all that is required to get started with the basic setup, use the OCM cli, and log into clusters with OCM Backplane with:

```
ocm-container --cluster-id <Cluster ID>
```

See the [Recommended Setup](#recommended-setup) section below for some opinionated defaults.

### Additional features

OCM Container has an additional feature set:

* Additional cluster environment variables
* (Legacy) AWS configuration and credential mounting
    * This pertains to AWS Profiles. New hires do not receive these anymore. These will eventually be phased out. 
* Mount Certificate Authorities from your host
* Google Cloud configuration and credential mounting
* JIRA CLI token and configuration
* OpsUtils directory mounting
* OSDCTL configuration mounting
* PagerDuty cli configuration mounting 
* Persistent Cluster Histories
* Persistent in-cluster image storage (podman only)
* ~/.bashrc personalization

Most features are enabled by default, though some may not do anything without additional configuration settings. Most features attempt to "intelligently" determine if they should be turned on based on whether or not the functionality is set up on your host. Features can be explicitly disabled or configured as desired. See [feature-specific documentation](docs/features) for any required settings.

## Usage

Running ocm-container can be done by executing the binary alone with no flags.

```bash
ocm-container
```

### Authentication

OCM authentication defaults to using your existing OCM Config, first looking for the `OCM_CONFIG` environment variable.

```bash
OCM_CONFIG="~/.config/ocm/ocm.json.prod" ocm-container
```

If no `OCM_CONFIG` is specified, ocm-container will login to the environment proved in the `OCMC_OCM_URL` environment variable (`prod`, `stage`, `int`, `prodgov`, `https://api.openshift.com`, etc) if set, then values provided by the `--ocm-url` flag.  If nothing is specified, the `--ocm-url` flag is set to "production" and that environment is used.

```bash
OCMC_OCM_URL=staging ocm-container

# or

ocm-container --ocm-url=staging
```

Upon login, OCM Container will copy a new ocm.json file to your `~/.config/ocm/` directory, in the format `ocm.json.ocm-container.$ocm_env`.  This file can be reused with the `OCM_CONFIG` environment variable in the future, if desired.

Passing a cluster ID to the command with `--cluster-id` or `-C` will log you into that cluster after the container starts. This can be the cluster's OCM UUID, the OCM internal ID or the cluster's display name.

### Cluster Login

```bash
ocm-container --cluster-id CLUSTER_ID
```

### Container engine options

Bind Mounts can be passed in the same format to ocm-container that you'd pass to `podman run`. ocm-container will check for the presence of a directory before attempting to bind it.

```bash
ocm-container -v "/path/to/my/dir:/dest/in/container:ro"
```

Additional container engine arguments can be passed to the container using the `--launch-ops` flag.  These will be passed as-is to the engine, and are a best-effort supported by ocm-container.

```bash
ocm-container --launch-opts "-v /tmp:/tmp:rw -e FOO=bar"
```

Some flags may conflict with ocm-container functionality.

## Flags, Environment and Configuration

Options for ocm-container can be passed as CLI flags or set as key: value pairs in ~/.config/ocm-container/ocm-container.yaml. 

For example, to set a specific ocm-container image tag rather than `latest`:

1. CLI Flag:  `ocm-container --image=ABCD`
2. Configuration File: `image: ABCD` to ~/.config/ocm-container/ocm-container.yaml, or `ocm-container configure set tag ABCD`

Configuration can be set manually in the configuration file.

The order of precedence is:

1. CLI Flags
2. Environment Variables
3. Configuration File

This means that if you have an option set in your Configuration File and then provide the flag to the next invocation, the value provided in the flag will be used over the Configuration File.

## Feature Set Configuration

All of the ocm-container feature sets are enabled by default, but some may require some additional configuration information passed (via CLI, ENV or configuration file, as show above) to actually do anything.

Every feature can be explicitly disabled. View the [feature-specific documentation](docs/features) for more information on each feature.

### Additional cluster environment variables

Automatically exports cluster-related environment variables when logging into a cluster with `--cluster-id`. These environment variables provide quick access to cluster metadata for use in scripts and commands within the container.

The following environment variables are automatically set:

* `CLUSTER_ID` - The internal OCM cluster ID
* `CLUSTER_UUID` - The external cluster UUID
* `CLUSTER_NAME` - The cluster's display name
* `CLUSTER_DOMAIN_PREFIX` - The cluster's domain prefix
* `CLUSTER_INFRA_ID` - The cluster's infrastructure ID
* `CLUSTER_HIVE_NAME` - The Hive cluster name (if available)

For HyperShift clusters, additional environment variables are set:

* `CLUSTER_MC_NAME` - The management cluster name
* `CLUSTER_SC_NAME` - The service cluster name
* `HCP_NAMESPACE` - The hosted control plane namespace
* `HC_NAMESPACE` - The short-form hosted cluster namespace
* `KUBELET_NAMESPACE` - The kubelet namespace

**Configuration:**

* No additional configuration required
* Requires `--cluster-id` to be specified when launching ocm-container
* [docs/features/additional-cluster-envs.md](/docs/features/addtional-cluster-envs.md)

### OpsUtils directory mounting

Red Hat SREs can mount the OPS Utils utilities into ocm-container, and can specify if the mount is read-only or read-write.

* Requires `ops_utils_dir: /fill/in/your/path/to/ops-sop/v4/utils` to be set
* [docs/features/ops-utils.md](/docs/features/ops-utils.md)

### OSDCTL configuration mounting

Mounts the [osdctl](https://github.com/openshift/osdctl) configuration directory (~/.config/osdctl) read-only into the container.

* No additional configuration required if your config file is in the expected location
* [docs/features/osdctl.md](/docs/features/osdctl.md)

### Persistent Cluster Histories

Stores cluster terminal history persistently on a per-cluster basis in directories in your ~/.config/ocm-container directory.

This feature is opt-in and is disabled by default. Follow instructions in [docs/features/persistent-histories.md](/docs/features/persistent-histories.md) to enable.

### ~/.bashrc personalization (or other)

Mounts a directory or a file (eg: ~/.bashrc or ~/.bashrc.d/, etc) from your host to ~/.config/personalizations.d (or ...personalizations.sh for a file) in the container.  You may specify if it is read-only or read-write.

* This feature is opt-in. Follow instruction in [docs/features/personalization.md](/docs/features/personalization.md).

## Micro, Minimal and Full container images

The `Containerfile` for ocm-container has three useful targets for building a "micro" image, a "minimal" image and the full-size ocm-container, each with additional tooling.  The `Makefile` has make targets to build each of these as well.

* micro: The micro image contains `ocm`, `backplane` and `oc`.  Makefile target: `make build-micro`
* minimal: The minimal image is build on the micro image, and adds all of the SRE [backplane tools](https://github.com/openshift/backplane-tools).  Makefile target: `make build-minimal`
* full:  The full ocm-container image builds on the minimal image and adds a number of other packages, tools, shell scripts and opinionated environment configuration (for example, to support auto-login to clusters, etc).  Makefile target: `make build`

## Personalize Your ocm-container

There are many options to personalize your ocm-container experience. For example, if you want to have your vim config passed in and available all the time, you could do something like this:

```bash
alias occ='ocm-container -v "/home/your_user/.vim:/root/.vim"'
```

Another common option is to have additional packages available that do not come in the standard build. You can create an additional Containerfile to run after you build the standard ocm-container build:

```dockerfile
FROM ocm-container:latest

RUN microdnf --assumeyes --nodocs update \
    && microdnf --assumeyes --nodocs install \
        lnav \
    && microdnf clean all \
    && rm -rf /var/cache/yum
```

_NOTE: When customizing ocm-container, use caution not to overwrite core tooling or default functionality in order to keep to the spirit of reproducible environments between SREs.  We offer the ability to customize your environment to provide the best experience, however the main goal of this tool is that all SREs have a consistent environment so that tooling "just works" between SREs._

## Recommended Setup

I personally recommend having an alias set up similar to the above example for OCM Container that automatically and explicitly passes the ocm url. Add this to your `bashrc` or `zshrc` or whatever shell configuration you may have:

```bash
alias occ='ocm-container --ocm-url production'
alias occs='ocm-container --ocm-url staging'
alias occi='ocm-container --ocm-url integration'
```

Then to log into a staging cluster I just run `occs -C my-cluster-name`, etc. 

## Advanced scripting with ocm-container

While a simple command could be run with the following, more advanced options are available if necessary.
```bash
while read -r cluster_id
do
    ocm-container -C $cluster_id -- oc version >> report.txt
```

For more advanced scripting needs, you can do something similar to the instructions below.

Given the following shell script saved on the local machine in ~/myproject/in-container.sh:

```bash
cat ~/myproject/in-container.sh
```

```text
#!/bin/bash

# get the version of the cluster
oc version >> report.txt
```

We can run that on-container with the following script which runs on the host (~/myproject/on-host.sh):

```bash
cat ~/myproject/on-host.sh
```

```text
#!/bin/bash

while read -r cluster_id
do
    echo "Cluster $cluster_id Information:" >> report.txt
    ocm-container -v "${HOME}/myproject/in-container.sh:/root/in-container.sh" -v "${HOME}/myproject/report.txt:/root/report.txt" -C $cluster_id -- exec /root/in-container.sh
    echo "----"
done < clusters.txt
```

Would loop through all clusters listed in `clusters.txt` and then run `oc version` on the cluster, and add the output into report.txt and then it would exit the container, and move to the next container and do the same.

## Troubleshooting

### SSH Config

If you're on a mac and you get an error similar to:

```text
Cluster is internal. Initializing Tunnel... /root/.ssh/config: line 34: Bad configuration option: usekeychain
```

you might need to add something similar to the following to your ssh config:

```bash
$ cat ~/.ssh/config | head
IgnoreUnknown   UseKeychain,AddKeysToAgent
Host *
  <snip>
  UseKeychain yes
```

UseKeychain is a MacOS specific directive which may cause issues on the linux container that ocm-container runs within.  Adding the `IgnoreUnknown UseKeychain` directive tells the ssh config to ignore that directive when it's unknown so it will not throw errors.

## Podman/M1 MacOS Instructions

The process is mostly the same. Assuming you have podman setup, with the following mounts on the podman machine:

```bash
brew install podman
podman machine init

# For podman versions less than 4.5.1, you need to manually pass in the home directory and /private directory mounts
# podman machine init -v ${HOME}:${HOME} -v /private:/private

podman machine start
```

Then you should just be able to build the container as usual

```bash
podman build -t ocm-container:latest .
```

### Note

When running local images you may need to set `pull` to `missing`

`ocm-container -I $IMAGE -t $TAG --pull missing`

If you are always running local or images with a set tag, you can set this in your config like this:

`ocm-container configure set pull missing`

_NOTE: the `ROSA` cli is not present on the arm64 version as there is no [pre-built arm64 binary](https://github.com/openshift/rosa/issues/874) that can be gathered, and we've decided that we don't use that cli enough to bother installing it from source within the build step._

## Development

The image for ocm-container is built nightly can by default are pulled from the registry at `quay.io/app-sre/ocm-container:latest`.  Alternatively you can build your own image and use it to invoke ocm-container with `make build-full-local`. 

## Continuous Integration

Continuous Integration log: [https://ci.int.devshift.net/blue/organizations/jenkins/openshift-ocm-container-gh-build-master/activity](https://ci.int.devshift.net/blue/organizations/jenkins/openshift-ocm-container-gh-build-master/activity)
