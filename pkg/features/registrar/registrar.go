package registrar

import (
	"github.com/openshift/ocm-container/pkg/features/jira"
	"github.com/openshift/ocm-container/pkg/features/pagerduty"
)

// the registrar package registers the various features by
// registering the features flag to disable them which invokes
// the init functions within the feature folders themselves

type flag struct {
	Name    string
	HelpMsg string
}

var featureFlags = []flag{
	{
		Name:    pagerduty.FeatureFlagName,
		HelpMsg: pagerduty.FlagHelpMessage,
	},
	{
		Name:    jira.FeatureFlagName,
		HelpMsg: jira.FlagHelpMessage,
	},
}

func FeatureFlags() []flag {
	return featureFlags
}
