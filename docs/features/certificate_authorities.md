# Certificate Authorities

By default, any CAs in the `/etc/pki/ca-trust/source/anchors` directory will be automatically mounted to the container in the same location. However, if you want to disable this functionality or use a custom folder path, the following configuation options are available:

```yaml
features:
  certificate_authorities:
    enabled: false
    source_anchors: /path/to/my/source/anchors
```
