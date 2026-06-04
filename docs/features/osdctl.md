# OSDCTL Configuration

This feature mounts your osdctl configuration into the container.

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

    # Mount options for the config file - either 'ro' (read-only) or 'rw' (read-write)
    # Defaults to 'ro'
    config_mount_options: ro
```

The config file will be mounted at `/root/.config/osdctl` inside the container.

## Vault Token

The `token_file` and `token_mount_options` settings are deprecated. Single-file bind mounts of `~/.vault-token` break vault's atomic rename during OIDC re-authentication inside the container.

Vault OIDC authentication is now handled via the [ports feature](ports.md), which dynamically maps port 8250 for the OIDC callback. osdctl manages the vault token in-process after authentication, avoiding the need to mount the token file.

## Notes

* Both absolute and `$HOME`-relative paths are supported for `config_file`
* If the config file is not found and you have provided configuration, the feature will exit with an error
* If you haven't provided any configuration for this feature, it will be silently disabled
