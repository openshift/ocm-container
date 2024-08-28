package ocmcontainer

import (
	"fmt"
	"regexp"
	"strconv"

	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/ocm-container/pkg/ocm"
	log "github.com/sirupsen/logrus"
)

const hiveRegex = `https?:\/\/api.([a-z0-9]+).[a-z0-9]+.[a-z0-9]+.openshiftapps.com:\d*`

type hcp struct {
	managementCluster string
	serviceCluster    string
	namespace         string
	hcpNamespace      string
	kubeletNamespace  string
}

type hive struct {
	name      string
	namespace string
}

type cluster struct {
	id         string
	uuid       string
	name       string
	baseDomain string
	hcp        *hcp
	hive       *hive
	env        map[string]string
}

// GetHcp returns the hcp struct for the given cluster
func GetHcp(conn *sdk.Connection, cluster *cmv1.Cluster) (*hcp, error) {
	// Get the HCP for the cluster
	hcp := &hcp{
		managementCluster: "",
		serviceCluster:    "",
		namespace:         "",
		hcpNamespace:      "",
		kubeletNamespace:  "",
	}
	if !cluster.Hypershift().Enabled() {
		return hcp, nil
	}

	resp, err := conn.ClustersMgmt().V1().Clusters().Cluster(cluster.ID()).Hypershift().Get().Send()
	if err != nil {
		return hcp, err
	}

	hcp.managementCluster = resp.Body().ManagementCluster()

	fltMgmtResp, err := conn.OSDFleetMgmt().V1().ManagementClusters().List().Parameter("search", fmt.Sprintf("name='%s'", hcp.managementCluster)).Send()
	if err != nil {
		return hcp, err
	}

	if kind := fltMgmtResp.Items().Get(0).Parent().Kind(); kind == "ServiceCluster" {
		hcp.serviceCluster = fltMgmtResp.Items().Get(0).Parent().Name()
	}

	hcp.namespace = fmt.Sprintf(ocm.HiveNSMap[conn.URL()], "ocm", cluster.ID())
	hcp.hcpNamespace = fmt.Sprintf(ocm.HiveNSMap[conn.URL()]+"-%s", "hcp", cluster.ID(), cluster.Name())
	hcp.kubeletNamespace = fmt.Sprintf("kubelet-%s", cluster.ID())

	return hcp, nil
}

// GetHive returns the hive struct for the given cluster
func GetHive(conn *sdk.Connection, cluster *cmv1.Cluster) (*hive, error) {
	var hive = &hive{
		name:      "",
		namespace: "",
	}
	shardPath, err := conn.ClustersMgmt().V1().Clusters().Cluster(cluster.ID()).ProvisionShard().Get().Send()
	if err != nil {
		return hive, err
	}

	if shardPath != nil {
		r, err := regexp.Compile(hiveRegex)
		if err != nil {
			return hive, err
		}
		hive.name = r.FindStringSubmatch(shardPath.Body().HiveConfig().Server())[1]
		hive.namespace = fmt.Sprintf(ocm.HiveNSMap[conn.URL()], "uhc", cluster.ID())
	}

	return hive, nil
}

// populateClusterEnv populates the cluster.env map with the cluster's information
func populateClusterEnv(v1Cluster *cmv1.Cluster, hcpInfo *hcp, hiveInfo *hive) map[string]string {
	m := make(map[string]string)

	// Add some cluster info to the ENVs
	m["CLUSTER_ID"] = v1Cluster.ID()
	m["CLUSTER_UID"] = v1Cluster.ID()
	m["CLUSTER_NAME"] = v1Cluster.Name()
	m["CLUSTER_BASEDOMAIN"] = v1Cluster.DNS().BaseDomain()
	m["CLUSTER_UUID"] = v1Cluster.ExternalID()

	m["CLUSTER_CLOUD_PROVIDER"] = v1Cluster.CloudProvider().ID()
	if v1Cluster.CloudProvider().ID() == "aws" && v1Cluster.AWS() != nil {
		m["CLUSTER_PRIVATE_LINK"] = strconv.FormatBool(v1Cluster.AWS().PrivateLink())
		if v1Cluster.AWS().STS().RoleARN() != "" {
			m["CLUSTER_STS"] = "true"
		}
	}

	m["CLUSTER_CCS"] = strconv.FormatBool(v1Cluster.CCS().Enabled())

	switch {
	case (hive{}) != *hiveInfo:
		log.Debug("Hive cluster detected")
		m["HIVE_NAME"] = hiveInfo.name
		m["HIVE_NS"] = hiveInfo.namespace
	case (hcp{}) != *hcpInfo:
		log.Debug("HPC cluster detected")
		m["MC_ID"] = hcpInfo.managementCluster
		m["SC_ID"] = hcpInfo.serviceCluster
		m["HC_NS"] = hcpInfo.namespace
		m["HCP_NS"] = hcpInfo.hcpNamespace
		m["KUBELET_NS"] = hcpInfo.kubeletNamespace
	}

	return m
}

// getCluster returns the cluster with the given key.  This is copied wholesale from:
// https://github.com/openshift-online/ocm-cli/blob/main/pkg/cluster/cluster.go@b0b70d9925bf280db02c4b259105aabe65fba599
func getCluster(connection *sdk.Connection, key string) (cluster *cmv1.Cluster, err error) {
	// Prepare the resources that we will be using:
	subsResource := connection.AccountsMgmt().V1().Subscriptions()
	clustersResource := connection.ClustersMgmt().V1().Clusters()

	// Try to find a matching subscription:
	subsSearch := fmt.Sprintf(
		"(display_name = '%s' or cluster_id = '%s' or external_cluster_id = '%s')",
		key, key, key,
	)
	subsListResponse, err := subsResource.List().
		Search(subsSearch).
		Size(1).
		Send()
	if err != nil {
		//lint:ignore ST1005 This is a copy of the original code
		err = fmt.Errorf("Can't retrieve subscription for key '%s': %v", key, err)
		return
	}

	// If there is exactly one matching subscription then return the corresponding cluster:
	subsTotal := subsListResponse.Total()
	if subsTotal == 1 {
		sub := subsListResponse.Items().Slice()[0]
		status, ok := sub.GetStatus()
		subID, _ := sub.GetID()
		if !ok || (status != "Reserved" && status != "Active") {
			//lint:ignore ST1005 This is a copy of the original code
			err = fmt.Errorf("Cluster was %s, see `ocm get subscription %s` for details", status, subID)
			return
		}
		id, ok := sub.GetClusterID()
		if ok {
			var clusterGetResponse *cmv1.ClusterGetResponse
			clusterGetResponse, err = clustersResource.Cluster(id).Get().
				Send()
			if err != nil {
				//lint:ignore ST1005 This is a copy of the original code
				err = fmt.Errorf(
					"Can't retrieve cluster for key '%s': %v",
					key, err,
				)
				return
			}
			cluster = clusterGetResponse.Body()
			return
		}
	}

	// If there are multiple subscriptions that match the cluster then we should report it as
	// an error:
	if subsTotal > 1 {
		//lint:ignore ST1005 This is a copy of the original code
		err = fmt.Errorf(
			"There are %d subscriptions with cluster identifier or name '%s'",
			subsTotal, key,
		)
		return
	}

	// If we are here then no subscription matches the passed key. It may still be possible that
	// the cluster exists but it is not reporting metrics, so it will not have the external
	// identifier in the accounts management service. To find those clusters we need to check
	// directly in the clusters management service.
	clustersSearch := fmt.Sprintf(
		"id = '%s' or name = '%s' or external_id = '%s'",
		key, key, key,
	)
	clustersListResponse, err := clustersResource.List().
		Search(clustersSearch).
		Size(1).
		Send()
	if err != nil {
		//lint:ignore ST1005 This is a copy of the original code
		err = fmt.Errorf("Can't retrieve clusters for key '%s': %v", key, err)
		return
	}

	// If there is exactly one cluster matching then return it:
	clustersTotal := clustersListResponse.Total()
	if clustersTotal == 1 {
		cluster = clustersListResponse.Items().Slice()[0]
		return
	}

	// If there are multiple matching clusters then we should report it as an error:
	if clustersTotal > 1 {
		//lint:ignore ST1005 This is a copy of the original code
		err = fmt.Errorf(
			"There are %d clusters with identifier or name '%s'",
			clustersTotal, key,
		)
		return
	}

	// If we are here then there are no subscriptions or clusters matching the passed key:
	//lint:ignore ST1005 This is a copy of the original code
	err = fmt.Errorf(
		"There are no subscriptions or clusters with identifier or name '%s'",
		key,
	)
	return
}

// This is copied wholesale from: https://github.com/openshift-online/ocm-cli/blob/main/pkg/cluster/cluster.go@@b0b70d9925bf280db02c4b259105aabe65fba599
// Regular expression to used to make sure that the identifier or name given by the user is
// safe and that it there is no risk of SQL injection:
var clusterKeyRE = regexp.MustCompile(`^(\w|-)+$`)

// IsValidClusterKey returns true if the given string is a valid cluster key.
func isValidClusterKey(clusterKey string) bool {
	return clusterKeyRE.MatchString(clusterKey)
}
