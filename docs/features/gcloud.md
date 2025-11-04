# Google Cloud CLI Configuration

This assumes that google cloud is working on your host machine.

* No additional configuration required if the configuration directory is in the standard location (`$HOME/.config/gcloud`)
* Can be explicitly disabled with the `--no-gcloud` flag or with the following yaml in the ocm-container config file:

```yaml
features:
  gcloud:
    enabled: false
```

## Additional Settings

The following config options are provided for the gcloud functionality:

```yaml
features:
  gcloud:
    config_dir: /path/to/config/gcloud
    config_mount: ro
```

### Mount Options

The `config_mount` option controls how the gcloud configuration directory is mounted into the container. Valid values are:

- `ro` - Read-only (default)
- `rw` - Read-write
- `z` - SELinux private unshared label
- `Z` - SELinux private shared label
- `ro,z` - Read-only with SELinux private unshared label
- `ro,Z` - Read-only with SELinux private shared label
- `rw,z` - Read-write with SELinux private unshared label
- `rw,Z` - Read-write with SELinux private shared label
