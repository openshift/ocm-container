package registrar

import (
	certificateauthorities "github.com/openshift/ocm-container/pkg/features/certificate-authorities"
	"github.com/openshift/ocm-container/pkg/features/jira"
	legacyawscredentials "github.com/openshift/ocm-container/pkg/features/legacy-aws-credentials"
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
	{
		Name:    legacyawscredentials.FeatureFlagName,
		HelpMsg: legacyawscredentials.FlagHelpMessage,
	},
	{
		Name:    certificateauthorities.FeatureFlagName,
		HelpMsg: certificateauthorities.FlagHelpMessage,
	},
}

func FeatureFlags() []flag {
	return featureFlags
}
