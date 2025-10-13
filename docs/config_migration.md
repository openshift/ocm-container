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

### Feature Configurations

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

#### PagerDuty
By default, Pagerduty now looks for the PD token file at `~/.config/pagerduty-cli/config.json`. The following configuration yaml changes have been made:

```yaml
.no_pagerduty (bool) -> .features.pagerduty.enabled (bool)

.pagerduty_dir_rw (bool) -> .features.pagerduty.mount (iota 'rw'|'ro')
```
