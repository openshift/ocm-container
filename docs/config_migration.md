# v0.0.0 -> v1.0.0 Migration Guide

In the following guide we will walk through the configuration changes for each independent feature as well as the configuration options that are needed for ocm-container to work.

## Deprecations
The following functionality has been deprecated:

* Scratch Directory mounting
  * Use the additionalMounts configuration instead  
  ```
  volumes:
  - "/path/to/scratch/dir:/root/scratch"
  ```
* Cluster ID can no longer be passed as an argument
* `--exec` flag is deprecated
* `--entrypoint` flag is deprecated
* `-x` flag is deprecated, use `--log-level debug` instead
* `-v` flag functionality has changed - now it passes bind volumes similar to `podman run` - previously this set the log level to info
* `-e` flag functionality has changed - now it passes environment variables similar to `podman run` - previously this was an alias for the `--exec` flag.

## Functionality Changes

### Command Execution and Argument Handling

* **Positional cluster IDs are no longer supported**. You must now use the `--cluster-id` flag to specify a cluster:
  ```bash
  # Old (deprecated):
  ocm-container my-cluster-id
  ocm-container -- my-cluster-id oc get pods

  # New (required):
  ocm-container --cluster-id my-cluster-id
  ocm-container --cluster-id my-cluster-id -- oc get pods
  ```

* **The `--exec` flag has been deprecated**. Commands must now be passed after `--`:
  ```bash
  # Old (deprecated):
  ocm-container --exec "oc get pods"

  # New (required):
  ocm-container -- oc get pods
  ```

* **The `-e` flag is now exclusively for environment variables**. It no longer has dual functionality for exec commands and is used only to pass additional environment variables into the container.

* **The `-` placeholder is no longer needed**. Previously, you needed to use `-` as a placeholder when executing commands without a cluster:
  ```bash
  # Old:
  ocm-container -- - echo "hello"

  # New:
  ocm-container -- echo "hello"
  ```

### Executing Commands in Running Containers

* **New feature: Execute commands against existing clusters**. You can now run commands inside the container after it logs into a cluster, and the container will automatically stop when the command completes:
  ```bash
  # Execute a single command and exit
  ocm-container --cluster-id my-cluster -- oc get pods

  # The container will:
  # 1. Start and login to the cluster
  # 2. Execute your command with live output
  # 3. Automatically stop and clean up
  ```

* **Improved signal handling**. Commands executed with `--` now properly handle interrupt signals (Ctrl+C), ensuring clean shutdown and resource cleanup. This _will_ immediately exit the container, as if the command was running directly on your host. Interupt signals passed while NOT running a command will keep the container session active. Example:
    ```bash
    {user@my-laptop} $ ocm-container -C some-cluster -- long-running-command
    my command output here
    0 ... 1 ... 2 ... ^C
    {user@my-laptop} $
    ```
    versus:
    ```bash
    {user@my-laptop} $ ocm-container -C some-cluster
    [in-cluster-prompt] $ long-running-command
    my command output here
    0 ... 1 ... 2 ... 3 ... ^C
    [in-cluster-prompt] $
    ```
    - Exits with status code 130 on interrupt (matching standard shell behavior)

### Feature Configurations

#### Certificate Authorities
By default, any CAs in `/etc/pki/ca-trust/source/anchors` will be automatically mounted in the container in the same directory.

To disable the functionality or customize the ca source anchor path, the following configuration changes have been made:

```yaml
.no_certificate_authorities (bool) -> .features.certificate_authorities.enabled (bool)

.ca_source_anchors (directory path) -> .features.certificate_authorities.source_anchors
```

#### Google Cloud
By default, ocm-container will look for your existing `~/.config/gcloud` directory and mount it to the same path inside the container. If this folder does not exist and it is not explicitly configured, ocm-container will continue silently. If your folder has been explicitly configured, it will give you a warning (but will continue to function). 

One change during the migration made is the feature disable flag. `--no-gcp` has been changed to `--no-gcloud` to be consistent with naming schemes and configuration

To disable this functionality or change the config directory, the following configuration changes have been made:

```yaml
.no_gcloud (bool) -> .features.gcloud.enabled (bool)
```

#### JIRA
By default, the JIRA integration looks for both the `JIRA_API_TOKEN` and `JIRA_AUTH_TYPE` environment variables, as well as a configuration file located at `~/.jira/.config.yml`. If the env vars are present, the config file is located and mounted. If the config file is not present, the env vars are still loaded. If the `JIRA_API_TOKEN` env var is not present, nothing is loaded. If the `JIRA_API_TOKEN` env var is present but the `JIRA_AUTH_TYPE` env var is not, `JIRA_AUTH_TYPE` will default to `bearer` and the config file will be attempted to be found.

The following configuration yaml changes have been made:

```yaml
.no_jira (bool) -> .features.jira.enabled (bool)
```

#### Legacy AWS Credentials
By default, this functionality picks up your `~/.aws/credentials` and `~/.aws/config` files and mounts them in the container. If those files don't exist, this will just silently fail.

If you want to disable this functionality, the following configuration yaml changes have to be made:

```yaml
.no_aws (bool) -> .features.legacy_aws_credentials.enabled (bool)
```

#### Ops Utils
This feature mounts the ops-sop/v4/utils directory into the container at `/root/ops-utils`. Unlike the previous implementation, this feature must now be explicitly configured with a source directory path. The following configuration yaml changes have been made:

```yaml
.ops_utils_dir (directory path) -> .features.ops_utils.source_dir (directory path)

.ops_utils_dir_rw (bool) -> .features.ops_utils.mount_options (iota 'rw'|'ro')
```

Note: The `source_dir` configuration is required for the feature to work - there is no default value.

#### OSDCTL
This feature mounts your osdctl configuration file and optional vault token into the container. Unlike the previous implementation, this feature must now be explicitly configured with a config file path. The following configuration yaml changes have been made:

```yaml
.no_osdctl (bool) -> .features.osdctl.enabled (bool)
```

New configuration options:
```yaml
.features.osdctl.config_file (file path) - defaults to '.config/osdctl'
.features.osdctl.token_file (file path) - defaults to '.vault-token'
.features.osdctl.config_mount_options (iota 'rw'|'ro') - defaults to 'ro'
.features.osdctl.token_mount_options (iota 'rw'|'ro') - defaults to 'rw'
```

Note: The `config_file` must exist for the feature to work. The `token_file` is optional.

#### PagerDuty
By default, Pagerduty now looks for the PD token file at `~/.config/pagerduty-cli/config.json`. The following configuration yaml changes have been made:

```yaml
.no_pagerduty (bool) -> .features.pagerduty.enabled (bool)

.pagerduty_dir_rw (bool) -> .features.pagerduty.mount (iota 'rw'|'ro')
```

#### Personalization
This feature mounts a directory or file containing your bash personalizations into the container. This feature must be explicitly configured with a source path. The following configuration yaml changes have been made:

```yaml
.no_personalization (bool) -> .features.personalization.enabled (bool)

.personalization_file (path) -> .features.personalization.source (path)

.personalization_dir_rw (bool) -> .features.personalization.mount_options (iota 'rw'|'ro')
```

Note: The `source` path is required for the feature to work. The source can be either a directory or a single file.

#### Persistent Histories
This feature provides per-cluster persistent bash history. It maintains separate command histories for each cluster you work with. This feature must be explicitly enabled in the configuration:

```yaml
.enable_persistent_histories (bool) -> .features.persistent_histories.enabled (bool) - defaults to false
```

Note: This feature requires a cluster-id to be provided and will automatically create subdirectories for each cluster's history. The storage_dir can be either an absolute path or relative to $HOME.

#### Image Cache
This feature provides persistent container image caching to improve startup times. Previously handled by the `persistent-images` feature flag, this is now a configurable feature with the following changes:

```yaml
--no-persistent-images (flag) -> --no-image-cache
.no-persistent-images (bool) -> .features.image_cache.enabled (bool)

New configuration options:
.features.image_cache.storage_dir (directory path) - defaults to '.config/ocm-container/images'
```

Note: This feature is disabled by default (opt-in) and must be explicitly enabled. When enabled, the storage directory path must exist. The storage_dir can be either an absolute path or relative to $HOME.

#### Backplane
This feature provides automatic mounting of your backplane configuration file into the container. This is a new feature that was previously handled directly in the core container initialization. The following configuration is available:

```yaml
.features.backplane.enabled (bool) - defaults to true

New configuration options:
.features.backplane.config_file (file path) - defaults to '.config/backplane/config.json'
```

This feature is enabled by default and will:
* Look for the backplane config in priority order:
  1. `$BACKPLANE_CONFIG` environment variable (highest priority)
  2. `config_file` setting (if specified)
  3. `$HOME/.config/backplane/config.json` (default)
* Mount the config file to `/root/.config/backplane/config.json` inside the container
* Set the `BACKPLANE_CONFIG` environment variable inside the container

The feature can be disabled with the `--no-backplane` flag or by setting `enabled: false` in the configuration.

The `config_file` path can be either an absolute path or relative to `$HOME`.

Note: If the backplane config file does not exist, the feature will fail gracefully and the container will still start. This allows users without backplane to use ocm-container without errors.

#### Ports
This feature provides port forwarding capabilities for the ocm-container, allowing you to expose container ports to your host system. Previously, the console port functionality was hardcoded in the core container initialization. This has been refactored into a configurable feature with the following changes:

```yaml
--disable-console-port (flag) -> --no-ports
--no-console-port (flag) -> --no-ports
.disable_console_port (bool) -> .ports.enabled (bool)
```

New configuration options:
```yaml
.ports.enabled (bool) - defaults to true
.ports.console.enabled (bool) - defaults to true
.ports.console.port (int) - defaults to 9999
```

This feature is enabled by default and will:
* Register the console port (default: 9999) for forwarding from the container to the host
* Map the console port to allow access to web consoles or services running in the container (such as `ocm backplane console`)
* After container startup, inspect the container to determine the actual host port assigned
* Write the port mapping information to `/tmp/portmap` inside the container for reference

The feature can be disabled with the `--no-ports` flag or by setting `enabled: false` in the configuration.

Note: Port forwarding is not required for core functionality. If port feature initialization fails, the error will be logged but the container will still start. However, if the port has been bound on the container but it cannot be looked up via the container engine, this will cause an error and exit.
