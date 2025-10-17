package additionalclusterenvs

import (
	"fmt"
	"regexp"

	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/ocm-container/pkg/features"
	"github.com/openshift/ocm-container/pkg/ocm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Define any defaults here as constants
const (
	FeatureFlagName = "no-additional-cluster-envs"
	FlagHelpMessage = "Disables additional cluster environment variables functionality"
)

// Any internal config needed for the setup of the feature
type config struct {
	Enabled bool `mapstructure:"enabled"`
}

// This is where we want to set all of our config defaults. If
// the user doesn't explicitly NEED to set something, set it
// here for them and allow them to overwrite it.
func newConfigWithDefaults() *config {
	config := config{}
	config.Enabled = true
	return &config
}

// Validate is where any custom configuration validation logic
// lives. This is where you need to validate your user's input
func (cfg *config) validate() error {
	return nil
}

type Feature struct {
	config *config

	// If the user provided a configuration, set that here
	// in case we want to handle initialization errors
	// differently because of that
	userHasConfig bool
}

// Enabled is where we determine whether or not the feature
// is explicitly enabled if opt-in or disabled if opt-out.
func (f *Feature) Enabled() bool {
	if !viper.IsSet("cluster-id") {
		log.Debugf("additional-cluster-envs disabled due to no cluster-id")
		return false
	}
	if !f.config.Enabled {
		log.Debugf("additional-cluster-envs disabled via config")
		return false
	}
	if viper.IsSet(FeatureFlagName) {
		log.Debugf("additional-cluster-envs disabled via flag")
		return false
	}
	return f.config.Enabled && !viper.IsSet(FeatureFlagName)
}

// If this feature is required for the functionality of
// ocm-container OR if a configuration error will be
// catastrophic to our user's experience, set this to true.
// Otherwise, if we lose a convenience function but we should
// still allow the user to use the container, then set false.
// In almost all cases, this should be set to false.
func (f *Feature) ExitOnError() bool {
	return false
}

// We want to self-contain the configuration functionality separate
// from the initialization so that we can read in the enabled config
func (f *Feature) Configure() error {
	cfg := newConfigWithDefaults()

	if !viper.IsSet("features.additional_cluster_envs") {
		f.config = cfg
		return nil
	}

	f.userHasConfig = true
	err := viper.UnmarshalKey("features.additional_cluster_envs", &cfg)
	if err != nil {
		return err
	}

	f.config = cfg
	err = cfg.validate()
	if err != nil {
		return err
	}

	return nil
}

// Initialize is the feature that we use to create the OptionSet
// for a given feature. An OptionSet is how the ocm-container
// program knows what options to pass into the container create
// command in order for the individual feature to work properly
func (f *Feature) Initialize() (features.OptionSet, error) {
	opts := features.NewOptionSet()

	ocmClient := ocm.GetClient()

	clusterID := viper.GetString("cluster-id")

	cluster, err := ocm.GetCluster(ocmClient, clusterID)
	if err != nil {
		return opts, err
	}

	opts.AddEnvKeyVal("CLUSTER_ID", cluster.ID())
	opts.AddEnvKeyVal("CLUSTER_UUID", cluster.ExternalID())
	opts.AddEnvKeyVal("CLUSTER_NAME", cluster.Name())
	opts.AddEnvKeyVal("CLUSTER_DOMAIN_PREFIX", cluster.DomainPrefix())
	opts.AddEnvKeyVal("CLUSTER_INFRA_ID", cluster.InfraID())

	shard, err := ocmClient.ClustersMgmt().V1().Clusters().
		Cluster(cluster.ID()).
		ProvisionShard().
		Get().
		Send()
	if shard != nil && err == nil {
		shard := shard.Body().HiveConfig().Server()
		r, _ := regexp.Compile(`hive[\-a-z0-9]+`)
		hive := r.FindString(shard)
		if hive != "" {
			opts.AddEnvKeyVal("CLUSTER_HIVE_NAME", hive)
		}
	}

	// Parse Hypershift-related values
	mgmtClusterName, svcClusterName, hcpNamespace := findHyperShiftInfo(ocmClient, cluster)
	if mgmtClusterName != "" {
		opts.AddEnvKeyVal("CLUSTER_MC_NAME", mgmtClusterName)
	}
	if svcClusterName != "" {
		opts.AddEnvKeyVal("CLUSTER_SC_NAME", svcClusterName)
	}
	if hcpNamespace != "" {
		opts.AddEnvKeyVal("HCP_NAMESPACE", hcpNamespace)
		hcNamespaceRegex, _ := regexp.Compile(`ocm-[a-z0-9]+-[a-z0-9]+`)
		hcNS := hcNamespaceRegex.FindString(hcpNamespace)
		opts.AddEnvKeyVal("HC_NAMESPACE", hcNS)
		opts.AddEnvKeyVal("KUBELET_NAMESPACE", fmt.Sprintf("kubelet-%s", cluster.ID()))
	}

	return opts, nil
}

// If initialize fails, how should we handle the error? This
// allows you to customize what log level to use or how to
// clean up anything you need to.
func (f *Feature) HandleError(err error) {
	if f.userHasConfig {
		log.Warnf("Error initializing additional-cluster-envs functionality: %v", err)
	}
	log.Debugf("Error initializing additional-cluster-envs functionality: %v", err)
}

// findHyperShiftMgmtSvcClusters returns the name of a HyperShift cluster's management and service clusters.
// It essentially ignores error as these endpoint is behind specific permissions by returning empty strings when any
// errors are encountered, which results in them not being printed in the output.
func findHyperShiftInfo(conn *sdk.Connection, cluster *cmv1.Cluster) (string, string, string) {
	if !cluster.Hypershift().Enabled() {
		return "", "", ""
	}

	hypershiftResp, err := conn.ClustersMgmt().V1().Clusters().
		Cluster(cluster.ID()).
		Hypershift().
		Get().
		Send()
	if err != nil {
		return "", "", ""
	}

	hcpNamespace := hypershiftResp.Body().HCPNamespace()
	mgmtClusterName := hypershiftResp.Body().ManagementCluster()
	fmMgmtResp, err := conn.OSDFleetMgmt().V1().ManagementClusters().
		List().
		Parameter("search", fmt.Sprintf("name='%s'", mgmtClusterName)).
		Send()
	if err != nil {
		return mgmtClusterName, "", hcpNamespace
	}

	if kind := fmMgmtResp.Items().Get(0).Parent().Kind(); kind == "ServiceCluster" {
		return mgmtClusterName, fmMgmtResp.Items().Get(0).Parent().Name(), hcpNamespace
	}

	// Shouldn't normally happen as every management cluster should have a service cluster
	return mgmtClusterName, "", hcpNamespace
}

func init() {
	f := Feature{}
	features.Register("additional-cluster-envs", &f)
}
