package ocm

// This is a hacky package - all the OCM stuff _should_ be contained inside the container image,
// but the "persistent histories" makes use of OCM information that has to be available before
// creating the container

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift-online/ocm-common/pkg/ocm/config"
	"github.com/openshift-online/ocm-common/pkg/ocm/connection-builder"
	sdk "github.com/openshift-online/ocm-sdk-go"
	auth "github.com/openshift-online/ocm-sdk-go/authentication"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/osdctl/pkg/utils"
	log "github.com/sirupsen/logrus"
)

const (
	productionURL    = "https://api.openshift.com"
	stagingURL       = "https://api.stage.openshift.com"
	integrationURL   = "https://api.integration.openshift.com"
	productionGovURL = "https://api.openshiftusgov.com"

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

var shortUrl = map[string]string{
	productionURL:    "prod",
	stagingURL:       "stage",
	integrationURL:   "int",
	productionGovURL: "prodgov",
}

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

func New(ocmcOcmUrl string) (*Config, error) {
	c := &Config{}

	c.Env = make(map[string]string)

	ocmConfig, err := config.Load()
	if err != nil {
		return c, err
	}

	if ocmConfig == nil {
		ocmConfig = new(config.Config)
	}

	switch {
	case os.Getenv("OCM_CONFIG") != "":
		// ocmConfig.URL will already be set in this case
		log.Debug("using OCM environment from $OCM_CONFIG")
		if ocmcOcmUrl != "" {
			log.Warnf("both $OCM_CONFIG and $OCMC_OCM_URL (or --ocm-url) are set; defaulting to $OCM_CONFIG for OCM environment")
		}
	default:
		log.Info("using OCM environment from $OCMC_OCM_URL (or --ocm-url)")
		ocmConfig.URL = url(ocmcOcmUrl)
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

	connectionBuilder := connection.NewConnection().Config(ocmConfig).AsAgent("ocm-container").WithApiUrl(ocmConfig.URL)
	connection, err := connectionBuilder.Build()
	if err != nil {
		return c, fmt.Errorf("error creating OCM connection: %s", err)
	}

	accessToken, refreshToken, err := connection.Tokens()
	if err != nil {
		return c, fmt.Errorf("error getting OCM tokens: %s", err)
	}

	ocmConfig.AccessToken = accessToken
	ocmConfig.RefreshToken = refreshToken

	// Note, we're saving our own copy of the OCM config here, to prevent overriding
	ocmConfigLocation, err := save(ocmConfig)
	if err != nil {
		return c, fmt.Errorf("error saving copy of OCM config: %s", err)
	}

	c.Env["OCMC_EXTERNAL_OCM_CONFIG"] = ocmConfigLocation
	c.Env["OCMC_INTERNAL_OCM_CONFIG"] = "/root/.config/ocm/ocm.json"

	_, err = os.Stat(ocmConfigLocation)
	if os.IsNotExist(err) {
		return c, fmt.Errorf("OCM config file does not exist: %s", ocmConfigLocation)
	}

	return c, nil
}

// url takes a string in the form of urlAliases, and returns
// the actual OCM URL
func url(s string) string {
	return urlAliases[s]
}

// alias takes a string in the form of an OCM_URL, and returns
// a short alias
func alias(s string) string {
	return shortUrl[s]
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

// save takes a *config.Config and saves it to a file alongside the existing OCM config
// The path is the same as the existing OCM config, but the filename follows the convention:
// ocm.json.ocm-container.$ocm_env
func save(cfg *config.Config) (string, error) {
	file, err := config.Location()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(file)
	err = os.MkdirAll(dir, os.FileMode(0755))
	if err != nil {
		return "", fmt.Errorf("can't create directory %s: %v", dir, err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("can't marshal config: %v", err)
	}

	cachedConfig := dir + "/ocm.json.ocm-container." + alias(cfg.URL)

	err = os.WriteFile(cachedConfig, data, 0600)
	if err != nil {
		return "", fmt.Errorf("can't write file '%s': %v", file, err)
	}
	return cachedConfig, nil
}
