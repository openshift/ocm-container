# Ops Utils Configuration

This feature allows you to mount a directory containing operational utilities into the container.

* Must be explicitly configured with a source directory
* Can be disabled with the `--no-ops-utils` flag or with the following yaml in the ocm-container config file:

```yaml
features:
  ops_utils:
    enabled: false
```

## Configuration

The following config options are provided for the ops-utils functionality:

```yaml
features:
  ops_utils:
    # Path to the directory containing your ops utilities
    # This is required for the feature to work
    source_dir: /path/to/ops/utils

    # Mount options - either 'ro' (read-only) or 'rw' (read-write)
    # Defaults to 'ro'
    mount_options: ro
```

The utilities will be mounted at `/root/ops-utils` inside the container.
