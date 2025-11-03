# PagerDuty token and configuration

Mounts the ~/.config/pagerduty-cli/config.json token file into the container.

* No additional configuration required, other than on first-run (see below)
* Can be disabled with the `--no-pagerduty` flag or with the following yaml in the config file:
```
features:
  pagerduty:
    enabled: false
```

## Initial Setup
In order to set up the Pagerduty CLI the first time, ensure that the config file exists first with `mkdir -p ~/.config/pagerduty-cli && touch ~/.config/pagerduty-cli/config.json`. You'll also need to mount the Pagerduty config file as writeable by setting the `config_mount: rw` configuration the first time. Once you've logged in to ocm-container, run `pd login` to do the initial setup.

You may then set the pagerduty feature config to be read-only on subsequent runs of ocm-container with:

```yaml
features:
  pagerduty:
    config_mount: ro
```

## Mount Options

The `config_mount` option controls how the PagerDuty configuration file is mounted into the container. Valid values are:

- `ro` - Read-only
- `rw` - Read-write (default)
- `z` - SELinux private unshared label
- `Z` - SELinux private shared label
- `ro,z` - Read-only with SELinux private unshared label
- `ro,Z` - Read-only with SELinux private shared label
- `rw,z` - Read-write with SELinux private unshared label
- `rw,Z` - Read-write with SELinux private shared label
