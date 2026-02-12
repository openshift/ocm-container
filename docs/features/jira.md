# JIRA CLI tool configuration

This assumes that the JIRA CLI tool is already set up and working outside of the container.

* No additional configuration required if the configuration file is in the standard location ($HOME/.jira/.config.yml)
* Additionally expects the `JIRA_API_TOKEN` and `JIRA_AUTH_TYPE` env vars to be set.
* Can be explicitly disabled with the `--no-jira` flag or with the following yaml in the ocm-container config file:
```
features:
  jira:
    enabled: false
```

## Additional Settings

The following config options are provided for the JIRA functionality:

```yaml
features:
  jira:
    config_file: /path/to/.jira/.config.yml
    config_mount: ro
```

### Mount Options

The `config_mount` option controls how the JIRA configuration file is mounted into the container. Valid values are:

- `ro` - Read-only (default)
- `rw` - Read-write
- `z` - SELinux private unshared label
- `Z` - SELinux private shared label
- `ro,z` - Read-only with SELinux private unshared label
- `ro,Z` - Read-only with SELinux private shared label
- `rw,z` - Read-write with SELinux private unshared label
- `rw,Z` - Read-write with SELinux private shared label
