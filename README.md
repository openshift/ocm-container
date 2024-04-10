# ocm-container

A quick environment for accessing OpenShift v4 clusters.

## Features
* Uses ephemeral containers per cluster login, keeping `.kube` configuration and credentials separate.
* Credentials are destroyed on container exit (container has `--rm` flag set)
* Displays current cluster-name, and OpenShift project (`oc project`) in bash PS1
* Infinitely extendable! Create your own Containerfile and reference `FROM: ocm-container:latest` and add whatever binaries you want on top

## Installation

First, download the latest release for your OS/Architecture: [https://github.com/openshift/ocm-container/releases](https://github.com/openshift/ocm-container/releases)

Setup the base configuration, setting your preferred container engine (Podman or Docker) and OCM Token:

```
ocm-container configure set engine CONTAINER_ENGINE
ocm-container configure set offline_access_token OCM_OFFLINE_ACCESS_TOKEN
```

__Note:__ the OCM offline_access_token will be deprecated in the near future. OCM Container will be updated to handle this and assist in migrating your configuration.

This is all that is required to get started with the basic setup, use the OCM cli, and log into clusters with OCM Backplane.

### Additional features

OCM Container has an additional feature set:

* AWS configuration and credential mounting from your host
* Mount Certificate Authorities from your host
* Google Cloud configuration and credential mounting from you host
* JIRA CLI token and configuration
* OpsUtils directory mounting
* OSDCTL configuration mounting
* PagerDuty token
* Persistent Cluster Histories
* ~/.bashrc personalization
* Scratch directory mounting

All features are enabled by default, though some may not do anything without additional configuration settings. Features can be disabled as desired. See feature-specific documentation below for any required settings.

## Usage

Running ocm-container can be done by executing the binary alone with no flags.

```
ocm-container
```

Passing a cluster ID to the command with `--cluster-id` will log you into that cluster after the container starts. This can be the cluster's OCM UUID, the OCM internal ID or the cluster's display name.

```
ocm-container --cluster-id CLUSTER_ID
```

By default, the container's Entrypoint is `/bin/bash`. You may also use the `--entrypoint=<command>` flag to change the container's Entrypoint as you would with a container engine.  The ocm-container binary also treats trailing non-flag arguments as container CMD arguments, again similar to how a container engine does.  For example, to execute the `ls` command as the Entrypoint and the flags `-lah` as the CMD, you can run:

```
ocm-container --entrypoint=ls -- -lah
```

__NOTE:__ The `--` delimiter between ocm-container flags and the CMD arguments.

You may also change the Entrypoint and CMD for use with an initial cluster ID for login, but note you will need to handle any OCM/Cluster login yourself:

```
ocm-container --entrypoint=ls --cluster-id CLUSTER_ID -- -lah
```

Additional container engine arguments can be passed to the container using the `--launch-ops` flag.  These will be passed as-is to the engine, and are a best-effort support.  Some flags may conflict with ocm-container function.

```
ocm-container --launch-opts "-v /tmp:/tmp:rw -e FOO=bar"
```

## Flags, Environment and Configuration

Options for ocm-container can be passed as CLI flags, environment variables prefixed with `OCMC_`, or set as key: value pairs in ~/.config/ocm-container/ocm-container.yaml.  The order of precedence is:

1. CLI Flags
2. Environment Variables
3. Configuration File

For example, to set a specific ocm-container image tag rather than `latest`:

1. CLI Flag:  `ocm-container --tag=ABCD`
2. Environment Variable: `OCMC_TAG=ABCD ocm-container` or `export OCMC_TAG=ABCD ; ocm-container` (etc.. according to your shell)
3. Configuration File: `tag: ABCD` to ~/.config/ocm-container/ocm-container.yaml, or `ocm-container configure set tag ABCD`

Configuration can be set manually in the configuration file, or set as key/value pairs by running `ocm-container configure set KEY VALUE`.

### Migrating configuration from the bash-based ocm-container.sh and env.source files

Users of ocm-container's original bash-based `ocm-container.sh` can migrate easily to the new Go binary.

Some things to note:

* You no longer need to clone the git repository
* You no longer need to build the image yourself (though you may, see "Development" below)
* The ocm-container bash alias is no longer needed - just execute the binary directly in your $PATH
* The ~/.config/ocm-container/env.source file has been replaced with ~/.config/ocm-container/ocm-container.yaml, a [Viper configuration file](https://github.com/spf13/viper)

Users of ocm-container's Go binary may import the existing configuration from ~/.config/ocm-container/env.source using the `ocm-container configure init` command for an interactive configuration setup:

```
ocm-container configure init
```

Or optionally, use the `--assume-yes` flag for a best-effort attempt to import the values:

```
ocm-container configure init --assume-yes
```

You can view the configuration in use with the `ocm-container configure get` subcommand:

```
ocm-container configure get
```

Example:

```
$ ocm-container configure get
Using config file: /home/chcollin/.config/ocm-container/ocm-container.yaml
engine: podman
offline_access_token: REDACTED
persistent_cluster_histories: false
repository: chcollin
scratch_dir: /home/chcollin/Projects
ops_utils_dir: /home/chcollin/Projects/github.com/openshift/ops-sop/v4/utils/
```

Sensitive values are set to REDACTED by default, and can be viewed by adding `--show-sensitive-values`.

## Feature Set Configuration:

All of the ocm-container feature sets are enabled by default, but some may require some additional configuration information passed (via CLI, ENV or configuration file, as show above) to actually do anything.

Every feature can be disabled by adding `--no-FEATURENAME` or setting `no_featurename: true` in the configuration file, etc.

### AWS configuration and credential mounting from your host

Mounts your ~/.aws/credentials and ~/.aws/config files read-only into ocm-container for use with the AWS CLI.

* No additional configuration required
* Can be disabled with `no_aws: true`

### Mount Certificate Authorities from your host

Mounts additional certificate authority trust bundles from a directory on your host and adds it to the bundle in ocm-container at /etc/pki/ca-trust/source/anchors, read-only.

* Requires `ca_source_anchors: PATH_TO_YOUR_CA_ANCHORS_DIRECTORY` to be set
* Can be disabled with `no_certificate_authorities: true`

### Google Cloud configuration and credential mounting from you host

Mounts Google Cloud configuration and credentials from ~/.config/gcloud on your host inside ocm-container, read only.

* No additional configuration required
* Can be disabled with `no_gcloud: true`

### JIRA CLI token and configuration

Mounts your JIRA token and config directory from ~/.config/.jira/token.json on your host read-only into ocm-container, and sets the `JIRA_API_TOKEN` and `JIRA_AUTH_TYPE=bearer` environment variables to be used with the JIRA CLI tool.

* No additional configuration required, other than on first-run (see below):
* Can be disabled with `no_jira: true`

Generate a Personal Access Token by logging into JIRA and clicking your user icon in the top right of the screen, and selecting "Profile". Then Navigate to "Personal Access Tokens" in the left sidebar, and generate a token.

If this is your first time using the JIRA CLI, ensure that the config file exists first with `mkdir -p ~/.config/pagerduty-cli && touch ~/.config/pagerduty-cli/config.json`. You'll also need to mount the JIRA config file as writeable by setting the `jira_dir_rw: true` configuration (or `export OCMC_JIRA_DIR_RW: true`) the first time. Once you've logged in to ocm-container, run `jira init` to do the initial setup.

You may then remove `jira_dir_rw: true` on subsequent runs of ocm-container.

### OpsUtils directory mounting

Red Hat SREs can mount the OPS Utils utilities into ocm-container, and can specify if the mount is read-only or read-write.

* Requires `ops_utils_dir: PATH_TO_YOUR_OPS_UTILS_DIRECTORY` to be set
* Optionally accepts `ops_utils_dir_rw: true` to enable read-write access in the mount
* Can be disabled with `no_ops_utils: true`

### OSDCTL configuration mounting

Mounts the [osdctl](https://github.com/openshift/osdctl) configuration directory (~/.config/osdctl) read-only into the container.

* No additional configuration required
* Can be disabled with `no_osdctl: true`

### PagerDuty token and configuration

Mounts the ~/.config/pagerduty-cli/config.json token file into the container.

* No additional configuration required, other than on first-run (see below)
* Can be disabled with `no_pagerduty: true`

In order to set up the Pagerduty CLI the first time, ensure that the config file exists first with `mkdir -p ~/.config/pagerduty-cli && touch ~/.config/pagerduty-cli/config.json`. You'll also need to mount the Pagerduty config file as writeable by setting the `pagerduty_dir_rw: true` configuration (or `export OCMC_PAGERDUTY_DIR_RW: true`) the first time. Once you've logged in to ocm-container, run `pd login` to do the initial setup.

You may then remove `pagerduty_dir_rw: true` on subsequent runs of ocm-container.

### Persistent Cluster Histories

Stores cluster terminal history persistently in directories in your ~/.config/ocm-container directory.

* Requires `enable_persistent_histories: true`; but this is toggle deprecated and will be removed in the future
* Otherwise no additional configuration required
* Can be disabled with `no_persistent_histories: true`

### ~/.bashrc personalization (or other)

Mounts a directory or a file (eg: ~/.bashrc or ~/.bashrc.d/, etc) from your host to ~/.config/personalizations.d (or ...personalizations.sh for a file) in the container.  You may specify if it is read-only or read-write.

* Requires `personalization_file: PATH_TO_FILE_OR_DIRECTORY_TO_MOUNT`
* Optionally, `personalization_dir_rw: true` can be set to make the mount read-write
* Can be disabled with `no_personalization: true`

### Scratch directory mounting

Mounts an arbitrary directory from your host to ~/scratch. You may specific if it is read-only or read-write.

* Requires `scratch_dir: PATH_TO_YOUR_SCRATCH_DIR`
* Optionally, `scratch_dir_rw: true` can be set to make the mount read-write
* Can be disabled with `no_scratch: true`


## Personalize Your ocm-container

There are many options to personalize your ocm-container experience. For example, if you want to have your vim config passed in and available all the time, you could do something like this:

``` sh
alias occ=`ocm-container -o "-v /home/your_user/.vim:/root/.vim"`
```

Another common option is to have additional packages available that do not come in the standard build. You can create an additional Containerfile to run after you build the standard ocm-container build:

``` dockerfile
FROM ocm-container:latest

RUN microdnf --assumeyes --nodocs update \
    && microdnf --assumeyes --nodocs install \
        lnav \
    && microdnf clean all \
    && rm -rf /var/cache/yum
```

```
NOTE: When customizing ocm-container, use caution not to overwrite core tooling or default functionality in order to keep to the spirit of reproducible environments between SREs.  We offer the ability to customize your environment to provide the best experience, however the main goal of this tool is that all SREs have a consistent environment so that tooling "just works" between SREs.
```

## Advanced scripting with ocm-container
We've recently added the ability to run a script within the container so that you can run ocm-container within a script.

Given the following shell script saved on the local machine in ~/myproject/in-container.sh:
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

```
Cluster is internal. Initializing Tunnel... /root/.ssh/config: line 34: Bad configuration option: usekeychain
```
you might need to add something similar to the following to your ssh config:

```
$ cat ~/.ssh/config | head
IgnoreUnknown   UseKeychain,AddKeysToAgent
Host *
  <snip>
  UseKeychain yes
```

UseKeychain is a MacOS specific directive which may cause issues on the linux container that ocm-container runs within.  Adding the `IgnoreUnknown UseKeychain` directive tells the ssh config to ignore that directive when it's unknown so it will not throw errors.

## Podman/M1 MacOS Instructions
The process is mostly the same. Assuming you have podman setup, with the following mounts on the podman machine:

```
brew install podman
podman machine init -v ${HOME}:${HOME} -v /private:/private
podman machine start
```
Then you should just be able to build the container as usual

```
podman build -t ocm-container:latest .
```

Note: the `ROSA` cli is not present on the arm64 version as there is no [pre-built arm64 binary](https://github.com/openshift/rosa/issues/874) that can be gathered, and we've decided that we don't use that cli enough to bother installing it from source within the build step.

## Development

The image for ocm-container is built nightly can by default are pulled from the registry at `quay.io/app-sre/ocm-container:latest`.  Alternatively you can build your own image and use it to invoke ocm-container.

__NOTE:__ This feature is currently experimental, and requires the [ocm-container Github repository](https://github.com/openshift/ocm-container) to be cloned, and for `make` to be installed on the system.  It currently uses the `make build` target.

Building a new image can be done with the `ocm-container build` command. The command accepts `--image` and `--tag` flags to name the resulting image:

```
ocm container build --image IMAGE --tag TAG
```
The resulting image would be in the naming convention: "IMAGE:TAG"
