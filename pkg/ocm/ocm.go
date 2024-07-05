package ocm

// This is a hacky package - all the OCM stuff _should_ be contained inside the container image,
// but the "persistent histories" makes use of OCM information that has to be available before
// creating the container

import (
	"os"
	"path/filepath"

	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/osdctl/pkg/utils"
)

const (
	productionURL    = "https://api.openshift.com"
	stagingURL       = "https://api.stage.openshift.com"
	integrationURL   = "https://api.integration.openshift.com"
	productionGovURL = "https://api.openshiftusgov.com"
)

// supprotedUrls is a shortened list of the urlAliases, for the help message
// We actually support all the urlAliases, but that's too many for the help
var (
	SupportedUrls = []string{
		"prod",
		"stage",
		"int",
		"prodgov",
	}
)

var urlAliases = map[string]string{
	"production":     productionURL,
	"prod":           productionURL,
	"prd":            productionURL,
	productionURL:    productionURL,
	"staging":        stagingURL,
	"stage":          stagingURL,
	"stg":            stagingURL,
	stagingURL:       stagingURL,
	"integration":    integrationURL,
	"int":            integrationURL,
	integrationURL:   integrationURL,
	"productiongov":  productionGovURL,
	"prodgov":        productionGovURL,
	"prdgov":         productionGovURL,
	productionGovURL: productionGovURL,
}

type Error string

func (e Error) Error() string { return string(e) }

const (
	errInvalidOcmUrl = Error("the specified ocm-url is invalid: %s")
)

type Config struct {
	Env map[string]string
}

func New(ocmUrl string) (*Config, error) {
	c := &Config{}

	c.Env = make(map[string]string)

	c.Env["OCM_URL"] = url(ocmUrl)

	if c.Env["OCM_URL"] == "" {
		return c, errInvalidOcmUrl
	}
	return c, nil
}

// url takes a string in the form of urlAliases, and returns
// the actual OCM URL
func url(s string) string {
	return urlAliases[s]
}

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

// Finds the OCM Configuration file and returns the path to it
// Taken wholesale from	openshift-online/ocm-cli
func GetOCMConfigLocation() (string, error) {
	if ocmconfig := os.Getenv("OCM_CONFIG"); ocmconfig != "" {
		return ocmconfig, nil
	}

	// Determine home directory to use for the legacy file path
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(home, ".ocm.json")

	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// Determine standard config directory
		configDir, err := os.UserConfigDir()
		if err != nil {
			return path, err
		}

		// Use standard config directory
		path = filepath.Join(configDir, "/ocm/ocm.json")
	}

	return path, nil
}
