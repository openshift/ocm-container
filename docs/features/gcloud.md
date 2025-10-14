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
```
