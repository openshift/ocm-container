# Persistent Histories Configuration

This feature provides per-cluster persistent bash history, allowing you to maintain separate command histories for each cluster you work with.

This feature is opt-in and is disabled by default.

## Configuration

The following config options are provided for the persistent-histories functionality:

```yaml
features:
  persistent_histories:
    # Enable or disable persistent histories
    # Default: false, must be explicitly enabled
    enabled: true

    # Path to the storage directory where cluster histories will be saved
    # Defaults to '.config/ocm-container/per-cluster-persistent'
    # Can be either an absolute path or relative to $HOME
    # This directory must be created manually for this functionality to work
    storage_dir: .config/ocm-container/per-cluster-persistent
```

If using the default storage directory, you may need to `mkdir -p ~/.config/ocm-container/per-cluster-persistent` for this functionality to work.

## How It Works

When enabled and a cluster-id is provided:

1. ocm-container resolves the cluster ID from OCM using the provided cluster identifier (name, ID, or external ID)
2. Creates a subdirectory named after the cluster ID
3. Mounts this directory at `/root/.cluster-history` inside the container with read-write permissions
4. Sets the `HISTFILE` environment variable to `/root/.cluster-history/.bash_history`

This ensures that each cluster maintains its own independent bash history that persists across container sessions.

## Storage Location

The storage directory is resolved in the following order:

1. **Absolute path**: If `storage_dir` is an absolute path (e.g., `/home/user/cluster-histories`), it will be used directly
2. **HOME-relative path**: If `storage_dir` is a relative path (e.g., `.config/ocm-container/per-cluster-persistent`), it will be resolved relative to `$HOME`

The directory structure will look like:
```
$HOME/.config/ocm-container/per-cluster-persistent/
├── 1a2b3c4d5e6f7g8h9i0j/
│   └── .bash_history
├── 9i8h7g6f5e4d3c2b1a0/
│   └── .bash_history
└── ...
```

## Requirements

- The feature must be explicitly enabled in the configuration
- A cluster-id must be provided via the `--cluster-id` flag
- The storage directory must be accessible

## Notes

* If the storage directory cannot be found, the container will exit with an error
* Each cluster's history is isolated from other clusters
* The history files persist across container sessions, allowing you to maintain context for each cluster
* If no cluster-id is provided, the history will not be kept
