package registrar

import (
	"github.com/openshift/ocm-container/pkg/features/backplane"
	certificateauthorities "github.com/openshift/ocm-container/pkg/features/certificate-authorities"
	"github.com/openshift/ocm-container/pkg/features/gcloud"
	imagecache "github.com/openshift/ocm-container/pkg/features/image-cache"
	"github.com/openshift/ocm-container/pkg/features/jira"
	legacyawscredentials "github.com/openshift/ocm-container/pkg/features/legacy-aws-credentials"
	opsutils "github.com/openshift/ocm-container/pkg/features/ops-utils"
	"github.com/openshift/ocm-container/pkg/features/osdctl"
	"github.com/openshift/ocm-container/pkg/features/pagerduty"
	persistenthistories "github.com/openshift/ocm-container/pkg/features/persistent-histories"
	"github.com/openshift/ocm-container/pkg/features/personalization"
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
	{
		Name:    gcloud.FeatureFlagName,
		HelpMsg: gcloud.FlagHelpMessage,
	},
	{
		Name:    opsutils.FeatureFlagName,
		HelpMsg: opsutils.FlagHelpMessage,
	},
	{
		Name:    osdctl.FeatureFlagName,
		HelpMsg: osdctl.FlagHelpMessage,
	},
	{
		Name:    personalization.FeatureFlagName,
		HelpMsg: personalization.FlagHelpMessage,
	},
	{
		Name:    persistenthistories.FeatureFlagName,
		HelpMsg: persistenthistories.FlagHelpMessage,
	},
	{
		Name:    imagecache.FeatureFlagName,
		HelpMsg: imagecache.FlagHelpMessage,
	},
	{
		Name:    backplane.FeatureFlagName,
		HelpMsg: backplane.FlagHelpMessage,
	},
}

func FeatureFlags() []flag {
	return featureFlags
}
