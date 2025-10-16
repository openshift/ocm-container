# Backplane Configuration

This feature provides automatic mounting and configuration of your backplane config file into the ocm-container, enabling seamless use of backplane CLI tools within the container.

* Enabled by default
* Can be disabled with the `--no-backplane` flag or with the following yaml in the ocm-container config file:

```yaml
features:
  backplane:
    enabled: false
```

## Configuration

The following config options are provided for the backplane functionality:

```yaml
features:
  backplane:
    # Enable or disable backplane configuration mounting
    # Default: true
    enabled: true

    # Optional path to the backplane config file
    # Defaults to '.config/backplane/config.json'
    # Can be either an absolute path or relative to $HOME
    config_file: .config/backplane/config.json
```

## How It Works

When enabled, ocm-container will:

1. Look for the backplane configuration file in the following priority order:
   - The path specified in the `BACKPLANE_CONFIG` environment variable (highest priority)
   - The path specified in `config_file` setting
   - Default location: `$HOME/.config/backplane/config.json` (if no config is provided)

2. Mount the configuration file to `/root/.config/backplane/config.json` inside the container

3. Set the `BACKPLANE_CONFIG` environment variable inside the container to point to the mounted config file

This allows backplane CLI commands to work seamlessly inside the ocm-container without any additional configuration.

## Configuration Location

The backplane config file location is resolved in the following priority order:

1. **Environment variable** (highest priority): If `BACKPLANE_CONFIG` is set, that path will be used directly
2. **Config file setting**: If `config_file` is specified in the configuration, it will be used
   - Can be an absolute path (e.g., `/home/user/backplane/config.json`)
   - Can be relative to `$HOME` (e.g., `.config/backplane/custom-config.json`)
3. **Default location**: If no configuration is provided, `$HOME/.config/backplane/config.json` is used

The path resolution checks for files in this order:
- Absolute path first
- `$HOME`-relative path second

## Error Handling

If the backplane configuration file does not exist at the expected location:

* The feature will fail to initialize
* An error will be logged at the debug level (if no user config is set)
* An error will be logged at the warning level (if user explicitly enabled the feature)
* The container will still start (backplane is not required for core functionality)

## Benefits

* **Seamless backplane integration**: Use backplane commands inside the container without manual configuration
* **Automatic detection**: No need to manually specify config paths in most cases
* **Flexible configuration**: Support for custom config locations via environment variables

## Notes

* The backplane config file is mounted with read-write permissions
* Changes made to the config inside the container will persist to the host system
* This feature is enabled by default but will gracefully fail if no backplane config exists
* Users without backplane installed can disable this feature to avoid debug log messages
