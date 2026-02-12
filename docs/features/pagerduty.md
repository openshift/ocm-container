# PagerDuty token and configuration

Mounts the ~/.config/pagerduty/token.json token file into the container.

* No additional configuration required
* Can be disabled with the `--no-pagerduty` flag or with the following yaml in the config file:
```
features:
  pagerduty:
    enabled: false
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
