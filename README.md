# ocm-container

An standardized environment for accessing OpenShift v4 clusters.

## Features

* Uses ephemeral containers per cluster login, keeping `.kube` configuration and credentials separate.
* Credentials are destroyed on container exit (container has `--rm` flag set)
* Displays current cluster-name, and OpenShift project (`oc project`) in bash PS1
* Infinitely extendable! Create your own Containerfile and reference `FROM: ocm-container:latest` and add whatever binaries you want on top

## Installation

First, download the latest release for your OS/Architecture: [https://github.com/openshift/ocm-container/releases](https://github.com/openshift/ocm-container/releases)

Setup the base configuration, setting your preferred container engine (Podman or Docker):

```bash
ocm-container configure set engine CONTAINER_ENGINE
```

This is all that is required to get started with the basic setup, use the OCM cli, and log into clusters with OCM Backplane with:

```
ocm-container --cluster-id <Cluster ID>
```

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

```bash
ocm-container
```

### Authentication

OCM authentication defaults to using your OCM Config, first looking for the `OCM_CONFIG` environment variable.

```bash
OCM_CONFIG="~/.config/ocm/ocm.json.prod" ocm-container
```

If no `OCM_CONFIG` is specified, ocm-container will login to the environment proved in the `OCMC_OCM_URL` environment variable (prod, stage, int, prodgov) if set, then values provided by the `--ocm-url` flag.  If nothing is specified, the `--ocm-url` flag is set to "production" and that environment is used.

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

### Entrypoint

By default, the container's Entrypoint is `/bin/bash`. You may also use the `--entrypoint=<command>` flag to change the container's Entrypoint as you would with a container engine.  The ocm-container binary also treats trailing non-flag arguments as container CMD arguments, again similar to how a container engine does.  For example, to execute the `ls` command as the Entrypoint and the flags `-lah` as the CMD, you can run:

```bash
ocm-container --entrypoint=ls -- -lah
```

_NOTE: The standard `--` delimiter between ocm-container flags and the CMD arguments must be used._

You may also change the Entrypoint and CMD for use with an initial cluster ID for login, but note you will need to handle any OCM/Cluster login yourself:

```bash
ocm-container --entrypoint=ls --cluster-id CLUSTER_ID -- -lah
```

### Container engine options

Additional container engine arguments can be passed to the container using the `--launch-ops` flag.  These will be passed as-is to the engine, and are a best-effort supported by ocm-container.

```bash
ocm-container --launch-opts "-v /tmp:/tmp:rw -e FOO=bar"
```

Some flags may conflict with ocm-container functionality.

## Flags, Environment and Configuration

Options for ocm-container can be passed as CLI flags, environment variables prefixed with `OCMC_`, or set as key: value pairs in ~/.config/ocm-container/ocm-container.yaml. 

For example, to set a specific ocm-container image tag rather than `latest`:

1. CLI Flag:  `ocm-container --tag=ABCD`
2. Environment Variable: `OCMC_TAG=ABCD ocm-container` or `export OCMC_TAG=ABCD ; ocm-container` (etc.. according to your shell)
3. Configuration File: `tag: ABCD` to ~/.config/ocm-container/ocm-container.yaml, or `ocm-container configure set tag ABCD`

Configuration can be set manually in the configuration file, or set as key/value pairs by running `ocm-container configure set KEY VALUE`.

The order of precedence is:

1. CLI Flags
2. Environment Variables
3. Configuration File

This means that if you have an option set in your Configuration File and then provide the flag to the next invocation, the value provided in the flag will be used over the Configuration File.

### Migrating configuration from the bash-based ocm-container.sh and env.source files

Users of ocm-container's original bash-based `ocm-container.sh` can migrate easily to the new Go binary.

Some things to note:

* You no longer need to clone the git repository
* You no longer need to build the image yourself (though you may, see "Development" below)
* The ocm-container bash alias is no longer needed - just execute the binary directly in your $PATH
* The ~/.config/ocm-container/env.source file has been replaced with ~/.config/ocm-container/ocm-container.yaml, a [Viper configuration file](https://github.com/spf13/viper)

Users of ocm-container's Go binary may import the existing configuration from ~/.config/ocm-container/env.source using the `ocm-container configure init` command for an interactive configuration setup:

```bash
ocm-container configure init
```

Or optionally, use the `--assume-yes` flag for a best-effort attempt to import the values:

```bash
ocm-container configure init --assume-yes
```

You can view the configuration in use with the `ocm-container configure get` subcommand:

```bash
ocm-container configure get
```

Example:

```bash
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

## Feature Set Configuration

All of the ocm-container feature sets are enabled by default, but some may require some additional configuration information passed (via CLI, ENV or configuration file, as show above) to actually do anything.

Every feature can be disabled, either with `OCMC_NO_FEATURENAME: true` environment variables, the `--no-FEATURENAME` flag on the CLI or setting `no-featurename: true` in the configuration file (~/.config/ocm-container/ocm-container.yaml), substituting "FEATURENAME" for the feature you wish to disable, eg:

* Env Var: `OCMC_NO_JIRA: true`
* CLI Flag: `ocm-contianer --no-jira`
* Config file entry: `no-jira: true`

### AWS configuration and credential mounting from your host

Mounts your ~/.aws/credentials and ~/.aws/config files read-only into ocm-container for use with the AWS CLI.

* No additional configuration required
* Can be disabled with `no-aws: true` (set in the ocm-container.yaml file)

### Mount Certificate Authorities from your host

Mounts additional certificate authority trust bundles from a directory on your host and adds it to the bundle in ocm-container at /etc/pki/ca-trust/source/anchors, read-only.

* Requires `ca_source_anchors: PATH_TO_YOUR_CA_ANCHORS_DIRECTORY` to be set
* Can be disabled with `no-certificate-authorities: true` (set in the ocm-container.yaml file)

### Google Cloud configuration and credential mounting from you host

Mounts Google Cloud configuration and credentials from ~/.config/gcloud on your host inside ocm-container, read only.

* No additional configuration required
* Can be disabled with `no-gcloud: true` (set in the ocm-container.yaml file)

### JIRA CLI token and configuration

Mounts your JIRA token and config directory from ~/.config/.jira/token.json on your host read-only into ocm-container, and sets the `JIRA_API_TOKEN` and `JIRA_AUTH_TYPE=bearer` environment variables to be used with the JIRA CLI tool.

* No additional configuration required, other than on first-run (see below):
* Can be disabled with `no-jira: true` (set in the ocm-container.yaml file)

Generate a Personal Access Token by logging into JIRA and clicking your user icon in the top right of the screen, and selecting "Profile". Then Navigate to "Personal Access Tokens" in the left sidebar, and generate a token.

If this is your first time using the JIRA CLI, ensure that the config file exists first with `mkdir -p ~/.config/.jira && touch ~/.config/.jira/config.json`. You'll also need to mount the JIRA config file as writeable by setting the `jira_dir_rw: true` configuration (or `export OCMC_JIRA_DIR_RW: true`) the first time. Once you've logged in to ocm-container, run `jira init` to do the initial setup.

You may then remove `jira_dir_rw: true` on subsequent runs of ocm-container.

### OpsUtils directory mounting

Red Hat SREs can mount the OPS Utils utilities into ocm-container, and can specify if the mount is read-only or read-write.

* Requires `ops_utils_dir: /fill/in/your/path/to/ops-sop/v4/utils` to be set
* Optionally accepts `ops_utils_dir_rw: true` to enable read-write access in the mount
* Can be disabled with `no-ops-utils: true` (set in the ocm-container.yaml file)

### OSDCTL configuration mounting

Mounts the [osdctl](https://github.com/openshift/osdctl) configuration directory (~/.config/osdctl) read-only into the container.

* No additional configuration required
* Can be disabled with `no-osdctl: true` (set in the ocm-container.yaml file)

### Persistent Cluster Histories

Stores cluster terminal history persistently in directories in your ~/.config/ocm-container directory.

* Requires `enable_persistent_histories: true`; but this is toggle deprecated and will be removed in the future
* Otherwise no additional configuration required
* Can be disabled with `no-persistent-histories: true` (set in the ocm-container.yaml file)

### ~/.bashrc personalization (or other)

Mounts a directory or a file (eg: ~/.bashrc or ~/.bashrc.d/, etc) from your host to ~/.config/personalizations.d (or ...personalizations.sh for a file) in the container.  You may specify if it is read-only or read-write.

* Requires `personalization_file: PATH_TO_FILE_OR_DIRECTORY_TO_MOUNT`
* Optionally, `personalization_dir_rw: true` can be set to make the mount read-write
* Can be disabled with `no-personalization: true` (set in the ocm-container.yaml file)

### Scratch directory mounting

Mounts an arbitrary directory from your host to ~/scratch. You may specific if it is read-only or read-write.

* Requires `scratch_dir: PATH_TO_YOUR_SCRATCH_DIR`
* Optionally, `scratch_dir_rw: true` can be set to make the mount read-write
* Can be disabled with `no-scratch: true` (set in the ocm-container.yaml file)

## Micro, Minimal and Full container images

The `Containerfile` for ocm-container has three useful targets for building a "micro" image, a "minimal" image and the full-size ocm-container, each with additional tooling.  The `Makefile` has make targets to build each of these as well.

* micro: The micro image contains `ocm`, `backplane` and `oc`.  Makefile target: `make build-micro`
* minimal: The minimal image is build on the micro image, and adds all of the SRE [backplane tools](https://github.com/openshift/backplane-tools).  Makefile target: `make build-minimal`
* full:  The full ocm-container image builds on the minimal image and adds a number of other packages, tools, shell scripts and opinionated environment configuration (for example, to support auto-login to clusters, etc).  Makefile target: `make build`

The micro and minimal images are not pushed to Quay.io, but can be build using the Makefile targets.  The full image is built nightly and can be pulled from Quay.io directly, or build manually with the Makefile target.

## Personalize Your ocm-container

There are many options to personalize your ocm-container experience. For example, if you want to have your vim config passed in and available all the time, you could do something like this:

```bash
alias occ=`ocm-container -o "-v /home/your_user/.vim:/root/.vim"`
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

## Advanced scripting with ocm-container

We've recently added the ability to run a script within the container so that you can run ocm-container within a script.

Given the following shell script saved on the local machine in ~/myproject/in-container.sh:

```bash
cat ~/myproject/in-container.sh
```

```text
#!/bin/bash

# source this so we get all of the goodness of ocm-container
source /root/.bashrc

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
    ocm-container -o "-v ${HOME}/myproject/script.sh:/root/script.sh -v ${HOME}/myproject/report.txt:/root/report.txt" -e /root/script.sh $cluster_id
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

The image for ocm-container is built nightly can by default are pulled from the registry at `quay.io/app-sre/ocm-container:latest`.  Alternatively you can build your own image and use it to invoke ocm-container.

_NOTE: This feature is currently experimental, and requires the [ocm-container Github repository](https://github.com/openshift/ocm-container) to be cloned, and for `make` to be installed on the system.  It currently uses the `make build` target._

Building a new image can be done with the `ocm-container build` command. The command accepts `--image` and `--tag` flags to name the resulting image:

```bash
ocm container build --image IMAGE --tag TAG
```

The resulting image would be in the naming convention: "IMAGE:TAG"

## Continuous Integration

Continuous Integration log: [https://ci.int.devshift.net/blue/organizations/jenkins/openshift-ocm-container-gh-build-master/activity](https://ci.int.devshift.net/blue/organizations/jenkins/openshift-ocm-container-gh-build-master/activity)
