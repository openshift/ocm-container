# Beta -> 1.0 Configuration Migration Guide

In the following guide we will walk through the configuration changes for each independent feature as well as the configuration options that are needed for ocm-container to work.

### Feature Configurations

#### PagerDuty
By default, Pagerduty now looks for the PD token file at `~/.config/pagerduty-cli/config.json`. The following configuration yaml changes have been made:

```yaml
.no_pagerduty (bool) -> .features.pagerduty.enabled (bool)

.pagerduty_dir_rw (bool) -> .features.pagerduty.mount (iota 'rw'|'ro')
```
