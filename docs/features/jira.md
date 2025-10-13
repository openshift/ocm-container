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

```
features:
  jira:
    config_file: /path/to/.jira/.config.yml
    config_mount: ro
```
