# OSDCTL Configuration

This feature allows you to mount your osdctl configuration and vault token into the container.

* Must be explicitly configured with a config file path
* Can be disabled with the `--no-osdctl` flag or with the following yaml in the ocm-container config file:

```yaml
features:
  osdctl:
    enabled: false
```

## Configuration

The following config options are provided for the osdctl functionality:

```yaml
features:
  osdctl:
    # Path to the osdctl config file
    # This is required for the feature to work
    # Defaults to '.config/osdctl' (relative to $HOME if not absolute)
    config_file: .config/osdctl

    # Path to the vault token file
    # This is optional - the feature will work without it
    # Defaults to '.vault-token' (relative to $HOME if not absolute)
    token_file: .vault-token

    # Mount options for the config file - either 'ro' (read-only) or 'rw' (read-write)
    # Defaults to 'ro'
    config_mount_options: ro

    # Mount options for the token file - either 'ro' (read-only) or 'rw' (read-write)
    # Defaults to 'rw'
    token_mount_options: rw
```

The config file will be mounted at `/root/.config/osdctl` inside the container, and the vault token (if present) will be mounted at `/root/.vault-token`.

## Notes

* Both absolute and `$HOME`-relative paths are supported for `config_file` and `token_file`
* The vault token file is optional - if it doesn't exist, only the config file will be mounted
* If the config file is not found and you have provided configuration, the feature will exit with an error
* If you haven't provided any configuration for this feature, it will be silently disabled
