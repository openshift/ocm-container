# Image Cache Configuration

This feature provides persistent container image caching across ocm-container sessions, improving startup times by reusing previously pulled container images.

* Disabled by default, must be explicitly enabled
* Can be disabled with the `--no-image-cache` flag or with the following yaml in the ocm-container config file:

```yaml
features:
  image_cache:
    enabled: false
```

## Configuration

The following config options are provided for the image-cache functionality:

```yaml
features:
  image_cache:
    # Enable or disable image caching
    # Default: false, must be explicitly enabled
    enabled: true

    # Path to the storage directory where container images will be cached
    # Defaults to '.config/ocm-container/images'
    # Can be either an absolute path or relative to $HOME
    storage_dir: .config/ocm-container/images
```

## How It Works

When enabled:

1. ocm-container mounts a persistent storage directory to `/var/lib/containers/storage/` inside the container
2. This directory stores all pulled container images and their layers
3. On subsequent container sessions, previously pulled images are reused instead of being re-downloaded

This significantly reduces container startup time, especially when working with large container images or slow network connections.

## Storage Location

The storage directory is resolved in the following order:

1. **Absolute path**: If `storage_dir` is an absolute path (e.g., `/home/user/container-images`), it will be used directly
2. **HOME-relative path**: If `storage_dir` is a relative path (e.g., `.config/ocm-container/images`), it will be resolved relative to `$HOME`

The directory structure will contain container storage data managed by the container engine (podman/docker).

## Benefits

* **Faster startup times**: Avoid re-downloading container images on each session
* **Reduced network usage**: Images are cached locally and reused
* **Improved offline capability**: Previously pulled images remain available without network access

## Notes

* If the storage directory path exists but is not a directory (e.g., it's a file), the container will exit with an error
* The storage directory is mounted with read-write permissions
* This feature is opt-in (disabled by default) and must be explicitly enabled in your configuration
* The cached images are shared across all ocm-container sessions
* Disk space usage will increase over time as more images are cached
