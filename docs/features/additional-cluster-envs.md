# Additional Cluster Environment Variables

This feature automatically exports cluster-related environment variables when logging into a cluster with `--cluster-id`. These environment variables provide quick access to cluster metadata for use in scripts and commands within the container.

This feature is enabled by default and requires a cluster-id to be provided.

## Configuration

The following config options are provided for the additional-cluster-envs functionality:

```yaml
features:
  additional_cluster_envs:
    # Enable or disable additional cluster environment variables
    # Default: true
    enabled: true
```

## How It Works

When enabled and a cluster-id is provided:

1. ocm-container resolves the cluster information from OCM using the provided cluster identifier (name, ID, or external ID)
2. Queries the OCM API for cluster details and metadata
3. For HyperShift clusters, queries additional management cluster and service cluster information
4. Exports all gathered information as environment variables inside the container

This ensures that cluster metadata is readily available for scripting and automation without needing to query OCM repeatedly.

## Environment Variables

### Standard Cluster Variables

The following environment variables are automatically set for all clusters:

| Variable | Description | Example Value |
|----------|-------------|---------------|
| `CLUSTER_ID` | The internal OCM cluster ID | `1a2b3c4d5e6f7g8h9i0j` |
| `CLUSTER_UUID` | The external cluster UUID | `12345678-1234-1234-1234-123456789abc` |
| `CLUSTER_NAME` | The cluster's display name | `my-production-cluster` |
| `CLUSTER_DOMAIN_PREFIX` | The cluster's domain prefix | `my-prod` |
| `CLUSTER_INFRA_ID` | The cluster's infrastructure ID | `my-prod-a1b2c` |
| `CLUSTER_HIVE_NAME` | The Hive cluster name (if available) | `hive-production-01` |

### HyperShift-Specific Variables

For HyperShift-enabled clusters, additional environment variables are set:

| Variable | Description | Example Value |
|----------|-------------|---------------|
| `CLUSTER_MC_NAME` | The management cluster name | `hcp-mgmt-us-east-1` |
| `CLUSTER_SC_NAME` | The service cluster name | `hcp-svc-us-east-1` |
| `HCP_NAMESPACE` | The hosted control plane namespace | `ocm-production-abc123-def456` |
| `HC_NAMESPACE` | The short-form hosted cluster namespace | `ocm-abc123-def456` |
| `KUBELET_NAMESPACE` | The kubelet namespace | `kubelet-1a2b3c4d5e6f7g8h9i0j` |

## Usage Examples

### Basic Usage

Simply launch ocm-container with a cluster-id:

```bash
ocm-container --cluster-id my-cluster-name
```

Once inside the container, all environment variables are available:

```bash
# Inside the container
echo $CLUSTER_NAME
# Output: my-cluster-name

echo $CLUSTER_ID
# Output: 1a2b3c4d5e6f7g8h9i0j
```

### Scripting Example

Use these environment variables in your scripts:

```bash
# Inside the container
cat > cluster_info.sh <<'EOF'
#!/bin/bash

echo "Cluster Information:"
echo "==================="
echo "Name: $CLUSTER_NAME"
echo "ID: $CLUSTER_ID"
echo "UUID: $CLUSTER_UUID"
echo "Infra ID: $CLUSTER_INFRA_ID"

if [ -n "$CLUSTER_MC_NAME" ]; then
    echo ""
    echo "HyperShift Information:"
    echo "======================"
    echo "Management Cluster: $CLUSTER_MC_NAME"
    echo "Service Cluster: $CLUSTER_SC_NAME"
    echo "HCP Namespace: $HCP_NAMESPACE"
fi
EOF

chmod +x cluster_info.sh
./cluster_info.sh
```

### Automation Example

Use the variables for automated operations:

```bash
# Query cluster status using the cluster ID
ocm get cluster $CLUSTER_ID

# Access HyperShift management cluster
if [ -n "$CLUSTER_MC_NAME" ]; then
    echo "This is a HyperShift cluster on management cluster: $CLUSTER_MC_NAME"
    # Additional HyperShift-specific operations
fi
```

## Requirements

- A cluster-id must be provided via the `--cluster-id` flag
- Valid OCM credentials with access to the cluster
- For HyperShift variables, the cluster must be a HyperShift cluster and the user must have appropriate permissions to query management cluster information

## Notes

* If no cluster-id is provided, this feature will be automatically disabled and no environment variables will be set
* If the OCM API cannot be reached or the cluster cannot be found, the container will display an error message
* The HyperShift-specific environment variables are only set if the cluster is a HyperShift cluster
* Some HyperShift variables may not be set if the user lacks permissions to query management cluster information
* Environment variables are set during container initialization and do not update if the cluster state changes

## Disabling the Feature

To disable this feature, use any of the following methods:

**Via config file (`~/.config/ocm-container/ocm-container.yaml`):**
```yaml
features:
  additional_cluster_envs:
    enabled: false
```

**Via environment variable:**
```bash
OCMC_NO_ADDITIONAL_CLUSTER_ENVS=true ocm-container --cluster-id my-cluster
```

**Via CLI flag:**
```bash
ocm-container --no-additional-cluster-envs --cluster-id my-cluster
```

**Via config file (alternative syntax):**
```yaml
no-additional-cluster-envs: true
```
