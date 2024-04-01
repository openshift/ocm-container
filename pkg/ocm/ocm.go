package ocm

// This is a hacky package - all the OCM stuff _should_ be contained inside the container image,
// but the "persistent histories" makes use of OCM information that has to be available before
// creating the container

import (
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/osdctl/pkg/utils"
)

func NewClient() (*sdk.Connection, error) {
	ocmClient, err := utils.CreateConnection()
	if err != nil {
		return nil, err
	}

	return ocmClient, err
}

// GetCluster takes an *sdk.Connection and a cluster identifier string, and returns a *sdk.Cluster
// The string can be anything - UUID, ID, DisplayName
func GetCluster(ocmClient *sdk.Connection, key string) (*cmv1.Cluster, error) {
	cluster, err := utils.GetCluster(ocmClient, key)
	if err != nil {
		return nil, err
	}

	return cluster, err
}

// GetClusterId takes an *sdk.Connection and a cluster identifier string, and returns the cluster ID
func GetClusterId(ocmClient *sdk.Connection, key string) (string, error) {
	cluster, err := GetCluster(ocmClient, key)
	if err != nil {
		return "", err
	}

	return cluster.ID(), err
}
