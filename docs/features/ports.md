# Ports Configuration

This feature provides port forwarding capabilities for the ocm-container, allowing you to expose container ports to your host system and access services running inside the container.

* Enabled by default
* Can be disabled with the `--no-ports` flag or with the following yaml in the ocm-container config file:

```yaml
features:
  ports:
    enabled: false
```

## Configuration

The following config options are provided for the ports functionality:

```yaml
features:
  ports:
    # Enable or disable all port forwarding
    # Default: true
    enabled: true

    # Console port configuration
    console:
      # Enable or disable console port forwarding
      # Default: true
      enabled: true

      # container port to bind for console access
      # Default: 9999
      port: 9999
```

## How It Works

When enabled, ocm-container will:

1. Register configured ports for forwarding from the container to the host
2. Map the console port (default: 9999) to allow access to web consoles or services running in the container
3. After container startup, inspect the container to determine the actual host port assigned
4. Write the port mapping information to `/tmp/portmap` inside the container for reference

The port forwarding allows you to:
- Access web services or port-forwards running inside the container from your host browser

## Port Mapping

Currently, the ports feature supports the following named ports:

### Console Port

The console port is designed for accessing web-based consoles or services:

- **Default container port**: 9999
- **Purpose**: General-purpose port for the `ocm backplane console` command to access the cluster's OCP console
- **Port map file**: After container starts, the actual mapped port is written to `/tmp/portmap` inside the container

## Configuration Examples

### Default Configuration

Use default settings (console port 9999):

```yaml
features:
  ports:
    enabled: true
```

### Custom Console Port

Change the console port to 8080:

```yaml
features:
  ports:
    console:
      port: 8080
```

### Disable Console Port

Disable only the console port while keeping the feature enabled:

```yaml
features:
  ports:
    console:
      enabled: false
```

### Disable All Ports

Disable the entire ports feature:

```yaml
features:
  ports:
    enabled: false
```

Or use the command-line flag:

```bash
ocm-container --no-ports
```

## Usage Examples

### Cluster's Backplane Web Console access
Start the backplane console in the container:
```bash
# Inside the container
ocm backplane console &
```
After the console container initializes inside of ocm-container, you should see a message:
```
== Console is available at http://127.0.0.1:39485 ==
```

Navigate to this link in your host browser, and you're now viewing the console inside the container.

### Accessing a Web Service

If you start a web server inside the container on port 9999:

```bash
# Inside the container
python3 -m http.server 9999
```

From outside the container you'll need to find which port is being used by the container with `podman ps`.

You can access it from your host browser at:
```
http://localhost:[host port]
```

### Checking Port Mappings

Inside the container, you can check the actual mapped port:

```bash
cat /tmp/portmap
```

This file contains the host port that was assigned to the console port.

## Error Handling

If port initialization fails:

* The error will be logged at the debug level (if no user config is set)
* The error will be logged at the warning level (if user explicitly configured ports)
* The container will still start (port forwarding is not required for core functionality)

Common reasons for port mapping failures:
- Port already in use on the host system
- Insufficient permissions to bind to the requested port
- Network configuration issues with the container engine

## Future Enhancements

The ports feature is designed to be extensible. Future versions may include:

- Support for additional named ports (e.g., prometheus, alertmanager, thanos, etc.)
- Automatic port conflict resolution

## Benefits

* **Easy service access**: Access container services from your host without manual port mapping
* **Consistent configuration**: Named ports provide predictable access points
* **Flexible setup**: Customize ports per your environment's needs
* **Non-intrusive**: Failures may or may not prevent container startup

## Notes

* Port mappings are established when the container starts
* Changing port configuration requires restarting the container
* The actual host port assigned is randomly selected by the container engine
* Port forwarding uses TCP protocol
