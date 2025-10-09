# v0.0.0 -> v1.0.0 Migration Guide

In the following guide we will walk through the configuration changes for each independent feature as well as the configuration options that are needed for ocm-container to work.

## Deprecations
The following functionality has been deprecated:

* Scratch Directory mounting
  * Use the additionalMounts configuration instead  
  ```
  volumes:
  - "/path/to/scratch/dir:/root/scratch"
  ```

### Feature Configurations

#### PagerDuty
By default, Pagerduty now looks for the PD token file at `~/.config/pagerduty-cli/config.json`. The following configuration yaml changes have been made:

```yaml
.no_pagerduty (bool) -> .features.pagerduty.enabled (bool)

.pagerduty_dir_rw (bool) -> .features.pagerduty.mount (iota 'rw'|'ro')
```
