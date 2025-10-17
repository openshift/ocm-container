package ocm

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
	"github.com/openshift/ocm-container/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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

var client *sdk.Connection

// tracks whether or not we have set to defer the closing
var deferredClose bool

func New() (*Config, error) {
	c := &Config{}
	c.Env = make(map[string]string)

	// if we load the  config from the filesystem or the keychain
	// we want to save the refreshed tokens so we don't have to
	// log in over and over. However, if there is no config, we
	// don't want to save anything.
	saveOriginalConfig := true

	ocmConfig, err := config.Load()
	if err != nil {
		return c, err
	}

	// If we do not have a loaded config and it doesn't exist, warn
	// the user and build one so that they can log in.
	if ocmConfig == nil {
		if !viper.GetBool("features.ocm.ignore-login-warning") {

			log.Warning("OCM config doesn't exist. You will be prompted to log in every time this is run. To prevent future log-in prompts, run `ocm login`")
		}
		saveOriginalConfig = false
		log.Debugf("Creating new ocm config...")
		ocmConfig = new(config.Config)
	}

	log.Debugf("Ensuring ocm config has defaults")
	err = ensureConfigDefaults(ocmConfig)
	if err != nil {
		return c, err
	}

	log.Debugf("Ensuring ocm config is armed")
	ensureArmed(ocmConfig)

	agentString := fmt.Sprintf("ocm-container-%s", utils.Version)

	connectionBuilder := connection.NewConnection().Config(ocmConfig).AsAgent(agentString).WithApiUrl(ocmConfig.URL)
	connection, err := connectionBuilder.Build()
	if err != nil {
		return c, fmt.Errorf("error creating OCM connection: %s", err)
	}
	// set the global ocm client
	client = connection

	// Save the tokens so that they're fresh
	accessToken, refreshToken, err := connection.Tokens()
	if err != nil {
		return c, fmt.Errorf("error getting OCM tokens: %s", err)
	}

	ocmConfig.AccessToken = accessToken
	ocmConfig.RefreshToken = refreshToken

	// Save the default config with the refreshed tokens so that we
	// are not prompted to log in every time this is run again
	if saveOriginalConfig {
		config.Save(ocmConfig)
	}

	// Now we're saving our own copy of the OCM config here, to prevent overriding inside the container.
	// and let's ensure that we overwrite the URL for the container's config
	ocmurl, err := url(viper.GetString("ocm-url"))
	ocmConfig.URL = ocmurl
	ocmConfigLocation, err := saveForEnv(ocmConfig)
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
func url(s string) (string, error) {
	u, ok := urlAliases[s]
	if !ok {
		return "", errInvalidOcmUrl
	}
	return u, nil
}

// alias takes a string in the form of an OCM_URL, and returns
// a short alias
func alias(s string) string {
	return shortUrl[s]
}

func GetClient() *sdk.Connection {
	// return the existing client if it exists
	return client
}

// CloseClient allows the global ocm client to be closed.
func CloseClient() error {
	if client == nil {
		return nil
	}
	return client.Close()
}

// GetCluster takes an *sdk.Connection and a cluster identifier string, and returns a *sdk.Cluster
// The string can be anything - UUID, ID, DisplayName
func GetCluster(connection *sdk.Connection, key string) (cluster *cmv1.Cluster, err error) {
	// Prepare the resources that we will be using:
	subsResource := connection.AccountsMgmt().V1().Subscriptions()
	clustersResource := connection.ClustersMgmt().V1().Clusters()

	// Try to find a matching subscription:
	subsSearch := fmt.Sprintf(
		"(display_name = '%s' or cluster_id = '%s' or external_cluster_id = '%s') and "+
			"status in ('Reserved', 'Active')",
		key, key, key,
	)
	subsListResponse, err := subsResource.List().
		Search(subsSearch).
		Send()
	if err != nil {
		err = fmt.Errorf("Can't retrieve subscription for key '%s': %v", key, err)
		return
	}

	// If there is exactly one matching subscription then return the corresponding cluster:
	subsTotal := subsListResponse.Total()
	if subsTotal == 1 {
		id, ok := subsListResponse.Items().Slice()[0].GetClusterID()
		if ok {
			var clusterGetResponse *cmv1.ClusterGetResponse
			clusterGetResponse, err = clustersResource.Cluster(id).Get().
				Send()
			if err != nil {
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
		Send()
	if err != nil {
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
		err = fmt.Errorf(
			"There are %d clusters with identifier or name '%s'",
			clustersTotal, key,
		)
		return
	}

	// If we are here then there are no subscriptions or clusters matching the passed key:
	err = fmt.Errorf(
		"There are no subscriptions or clusters with identifier or name '%s'",
		key,
	)
	return
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
func saveForEnv(cfg *config.Config) (string, error) {
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

// ensureArmed validates that a given ocmConfig is "armed" and
// ensures that the credentials are valid and ready to be saved
// if the credentials are invalid or expired, it will initiate login
func ensureArmed(ocmConfig *config.Config) error {
	armed, reason, err := ocmConfig.Armed()
	if err != nil {
		return fmt.Errorf("error checking OCM config arming: %s", err)
	}

	var token string

	if !armed {
		log.Debugf("not logged into OCM: %s", reason)
		token, err = auth.InitiateAuthCode(ocmContainerClientId)
		if err != nil {
			return fmt.Errorf("error initiating auth code: %s", err)
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
			return fmt.Errorf("error parsing token: %s", err)
		}

		typ, err := config.TokenType(parsedToken)
		if err != nil {
			return fmt.Errorf("error determining token type: %s", err)
		}

		switch typ {
		case "Bearer", "":
			log.Debugf("token type is '%s'; assuming it is an AccessToken", typ)
			ocmConfig.AccessToken = token
		case "Refresh":
			log.Debugf("token type is '%s'; assuming it is a RefreshToken", typ)
			ocmConfig.AccessToken = ""
			ocmConfig.RefreshToken = token
		default:
			return fmt.Errorf("unknown token type: %s", typ)
		}
	}

	return nil
}

func ensureConfigDefaults(cfg *config.Config) error {
	if cfg.ClientID == "" {
		cfg.ClientID = ocmContainerClientId
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = sdk.DefaultTokenURL
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = defaultOcmScopes
	}
	if cfg.URL == "" {
		cfg.URL = productionURL
		if viper.IsSet("ocm-url") {
			sessionUrl, err := url(viper.GetString("ocm-url"))
			if err != nil {
				return err
			}
			cfg.URL = sessionUrl
		}
	}
	return nil
}
