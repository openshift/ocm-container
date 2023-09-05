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
* `aws`
* `oc`
* `ocm`
* `omg` 
* `osdctl`
* `pd`
* `rosa`
* `velero`
* `yq`

## Features:
* Does not mount any host filesystem objects as read/write, only uses read-only mounts, except for configurable mountpoints.
* Uses ephemeral containers per cluster login, keeping seperate `.kube` configuration and credentials.
* Credentials are destroyed on container exit (container has `--rm` flag set)
* Displays current cluster-name, and OpenShift project (`oc project`) in bash PS1
* Ability to login to private clusters without using a browser
* Ability to personalize it - `$PERSONALIZATION_FILE` can be set in `env.source` which automatically sources the file (or `.sh` files within a directory if PERSONALIZATION_FILE points to a directory) allowing for personal customizations
* Infinitely extendable:
  * Create your own Containerfile and reference `FROM: ocm-container:latest` and add whatever binaries you want on top
  * Mount as many other directories as you want with the `-o` flag (ex: want your vim config? `-o '-v /path/to/.vim:/root/.vim'`)

OCM Container also includes multiple scripts for your ease of use. For a quick overview of what is available, run `list-utils`.

## Installation:

* clone this repo
* `./init.sh` (repeat until all required fields are set)
* `./build.sh` (while this is building you can run the rest in another tab until the final command)
* optional: have local aws keys (put on ${HOME}/.aws)
* optional: have local gcp keys (put on ${HOME}/.config/gcloud)
* optional: add alias in `~/.bashrc`
  * `alias ocm-container-stg="OCM_URL=stg ocm-container"`
  * `alias ocm-container-local='OCM_CONTAINER_LAUNCH_OPTS="-v ${PWD}:/root/local -w /root/local" ocm-container'`
* optional: add your PagerDuty API token in `$HOME/.config/pagerduty-cli/config.json`
* Connect to the VPN
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

## Personalize it

There are many options to personalize your ocm-container experience. For example, if you want to have your vim config passed in and available all the time, you could do something like this:

``` sh
alias occ=`ocm-container -o "-v /home/myuser/.vim:/root/.vim"`
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
NOTE: When customizing ocm-container, use caution not to overwrite core tooling or default functionality in order to keep to the spirit of reproducable environments between SREs.  We offer the ability to customize your environment to provide the best experience, however the main goal of this tool is that all SREs have a consistent environment so that tooling "just works" between SREs.
```

### Pagerduty CLI setup

In order to set up the Pagerduty CLI the first time, ensure that the config file exists first with `mkdir -p ~/.config/pagerduty-cli && touch ~/.config/pagerduty-cli/config.json`.

Then, modify the mount inside [ocm-container.sh](https://github.com/openshift/ocm-container/blob/master/ocm-container.sh#L149) and remove the `:ro` flag at the end of the mount - this will allow the container the ability to write to the config file. 

Next, launch the container and then follow the instructions to log in with `pd login`.

Finally, undo the changes to the mount by re-adding the `:ro` flag.

### JIRA CLI setup

Create the JIRA configuration directory `mkdir -p ~/.config/.jira`.

Generate a Personal Access Token by logging into JIRA and clicking your user icon in the top right of the screen, and selecting "Profile". Then Navigate to "Personal Access Tokens" in the left sidebar, and generate a token.

If you wish to use the `jira` cli outside of ocm-container, add the following to your environment by adding it to your .zshrc or .bashrc file. If you only wish to use this within ocm-container, you can also add the following to your `.env.source` file:

```bash
JIRA_API_TOKEN="[insert your personal access token here]"
JIRA_AUTH_TYPE="bearer"
```

These env vars will automatically be picked up ocm-container. If this is your first time using the JIRA CLI, you'll also need to mount the [JIRA config file](https://github.com/openshift/ocm-container/blob/master/ocm-container.sh#L137) as writeable by removing the `:ro` flag at the end of the mount instructions. Once you've logged in and the CLI has been configured, you should be able to re-add the read-only flag.

## Automatic Login to a cluster:

```
ocm-container test-cluster
```

## Example:

```
$ ocm-container
[~ {production} ]$ sre-login test-cluster
Logging into cluster test-cluster
Cluster ID: 15uq6splkva07jsjwebn4890sph4vs3p8m
$ oc config current-context
default/test-cluster/test-user
```

### Automatic Login
We've built in functionality to simplify the cluster login steps.  Now within the container you can run `sre-login cluster-id` and it will refresh your ocm login, create a tunnel within the container if necessary, and then log-in to the cluster.

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
