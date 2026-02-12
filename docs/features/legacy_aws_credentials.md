# Legacy AWS Credentials

This functionality passes through your `~/.aws/credentials` and `~/.aws/config` files to the container. This does not block the startup of the container if these files do not exist.

To disable this functionality, you can pass the `--no-legacy-aws-credentials` flag or disable with the following config:

```yaml
features:
  legacy_aws_credentials:
    enabled: false
```
