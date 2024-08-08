package ocm

// This is a hacky package - all the OCM stuff _should_ be contained inside the container image,
// but the "persistent histories" makes use of OCM information that has to be available before
// creating the container

import (
	"fmt"
	"os"

	"github.com/openshift-online/ocm-cli/pkg/config"
	sdk "github.com/openshift-online/ocm-sdk-go"
	auth "github.com/openshift-online/ocm-sdk-go/authentication"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/openshift/osdctl/pkg/utils"
	log "github.com/sirupsen/logrus"
)

const (
	productionURL    = "https://api.openshift.com"
	stagingURL       = "https://api.stage.openshift.com"
	integrationURL   = "https://api.integration.openshift.com"
	productionGovURL = "https://api.openshiftusgov.com"

	ocmConfigDest      = "/root/.config/ocm/ocm.json"
	ocmConfigMountOpts = "ro" // This should stay read-only, to keep the container from impacting the external environment

	ocmContainerClientId = "ocm-cli"
)

// SupportedUrls is a shortened list of the urlAliases, for the help message
// We actually support all the urlAliases, but that's too many for the help
var (
	defaultOcmScopes = []string{"openid"}

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
	Env    map[string]string
	Mounts []engine.VolumeMount
}

func New(ocmUrl string) (*Config, error) {
	c := &Config{}

	c.Env = make(map[string]string)

	// OCM URL is required by the OCM CLI inside the container
	// otherwise the URL will be overridden by the saved OCM config
	c.Env["OCM_URL"] = url(ocmUrl)
	c.Env["OCMC_OCM_URL"] = url(ocmUrl)

	if c.Env["OCMC_OCM_URL"] == "" {
		return c, errInvalidOcmUrl
	}

	ocmConfig, err := config.Load()
	if err != nil {
		return c, err
	}

	if ocmConfig == nil {
		ocmConfig = new(config.Config)
	}

	armed, reason, err := ocmConfig.Armed()
	if err != nil {
		return c, fmt.Errorf("error checking OCM config arming: %s", err)
	}

	var token string

	if !armed {
		log.Debugf("not logged into OCM: %s", reason)
		token, err = auth.InitiateAuthCode(ocmContainerClientId)
		if err != nil {
			return c, fmt.Errorf("error initiating auth code: %s", err)
		}
	} else {
		log.Debug("already logged into OCM")
		token = ocmConfig.AccessToken
	}

	if config.IsEncryptedToken(token) {
		log.Debug("OCM token is encrypted; assuming it is a RefreshToken")
		ocmConfig.AccessToken = ""
		ocmConfig.RefreshToken = token
	} else {
		log.Debug("OCM token is not encrypted; assuming it is an AccessToken")

		parsedToken, err := config.ParseToken(token)
		if err != nil {
			return c, fmt.Errorf("error parsing token: %s", err)
		}

		typ, err := config.TokenType(parsedToken)
		if err != nil {
			return c, fmt.Errorf("error determining token type: %s", err)
		}

		switch typ {
		case "Bearer", "":
			log.Debug("token type is Bearer or empty; assuming it is an AccessToken")
			ocmConfig.AccessToken = token
		case "Refresh":
			log.Debug("token type is Refresh; assuming it is a RefreshToken")
			ocmConfig.AccessToken = ""
			ocmConfig.RefreshToken = token
		default:
			return c, fmt.Errorf("unknown token type: %s", typ)
		}

	}

	ocmConfig.ClientID = ocmContainerClientId
	ocmConfig.TokenURL = sdk.DefaultTokenURL
	ocmConfig.Scopes = defaultOcmScopes
	// note - purposely not setting the ocmConfig.URL here
	// to prevent overwriting the URL *outside* of the container
	// The gateway is set by the OCM_URL env inside the container.  See above.

	connection, err := ocmConfig.Connection()
	if err != nil {
		return c, fmt.Errorf("error creating OCM connection: %s", err)
	}

	accessToken, refreshToken, err := connection.Tokens()
	if err != nil {
		return c, fmt.Errorf("error getting OCM tokens: %s", err)
	}

	ocmConfig.AccessToken = accessToken
	ocmConfig.RefreshToken = refreshToken

	err = config.Save(ocmConfig)
	if err != nil {
		log.Warnf("non-fatal error saving OCM config: %s", err)
	}

	ocmConfigLocation, err := config.Location()
	if err != nil {
		return c, fmt.Errorf("unable to identify OCM config location: %s", err)
	}

	ocmVolume := engine.VolumeMount{
		Source:       ocmConfigLocation,
		Destination:  ocmConfigDest,
		MountOptions: ocmConfigMountOpts,
	}

	_, err = os.Stat(ocmVolume.Source)
	if !os.IsNotExist(err) {

		c.Mounts = append(c.Mounts, ocmVolume)
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
