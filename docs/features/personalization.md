# Personalization Configuration

This feature allows you to mount a directory or file containing your bash personalizations (like aliases, functions, etc.) into the container.

* Must be explicitly configured with a source path
* Can be disabled with the `--no-personalization` flag or with the following yaml in the ocm-container config file:

```yaml
features:
  personalization:
    enabled: false
```

## Configuration

The following config options are provided for the personalization functionality:

```yaml
features:
  personalization:
    # Path to the directory or file containing your personalizations
    # This is required for the feature to work
    # Can be either a directory or a single file
    source: /path/to/personalizations

    # Mount options - either 'ro' (read-only) or 'rw' (read-write)
    # Defaults to 'ro'
    mount_options: ro
```

## Mounting Behavior

- **If source is a directory**: It will be mounted at `/root/.config/personalizations.d/` inside the container
- **If source is a file**: It will be mounted at `/root/.config/personalizations.d/personalizations.sh` inside the container

The container's bashrc is configured to automatically source all `.sh` files in the `/root/.config/personalizations.d/` directory.

## Notes

* The source path is required for the feature to work
* If the source path doesn't exist and you have provided configuration, the feature will exit with an error
* If you haven't provided any configuration for this feature, it will be silently disabled
