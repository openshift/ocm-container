package persistenthistories

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/openshift/ocm-container/pkg/features"
	"github.com/openshift/ocm-container/pkg/ocm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

const (
	FeatureFlagName = "no-persistent-histories"
	FlagHelpMessage = "Disable persistent histories file mounts and environment"

	histFile          = ".bash_history"
	destDir           = "/root/.cluster-history"
	defaultStorageDir = ".config/ocm-container/per-cluster-persistent"
)

type config struct {
	Enabled    bool   `mapstructure:"enabled"`
	StorageDir string `mapstructure:"storage_dir"`
}

func newConfigWithDefaults() *config {
	config := config{}
	config.Enabled = false
	config.StorageDir = defaultStorageDir
	return &config
}

func (cfg *config) validate() error {
	return nil
}

type Feature struct {
	config *config

	userHasConfig bool
	afs           *afero.Afero

	criticalError bool
}

func (f *Feature) Enabled() bool {
	if !f.config.Enabled {
		log.Debugf("persistent-histories disabled via config")
		return false
	}
	if viper.IsSet(FeatureFlagName) {
		log.Debugf("persistent-histories disabled via flag")
		return false
	}
	// Only enable if cluster-id is provided
	if !viper.IsSet("cluster-id") || viper.GetString("cluster-id") == "" {
		log.Debugf("persistent-histories disabled: no cluster-id provided")
		return false
	}

	return true
}

func (f *Feature) ExitOnError() bool {
	return f.criticalError
}

func (f *Feature) Configure() error {
	cfg := newConfigWithDefaults()

	if !viper.IsSet("features.persistent_histories") {
		f.config = cfg
		return nil
	}

	f.userHasConfig = true
	err := viper.UnmarshalKey("features.persistent_histories", &cfg)
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

func (f *Feature) Initialize() (features.OptionSet, error) {
	opts := features.NewOptionSet()

	cluster := viper.GetString("cluster-id")
	if cluster == "" {
		return opts, fmt.Errorf("cluster-id is required for persistent histories")
	}

	// Get the cluster ID from OCM
	ocmClient, err := ocm.NewClient()
	if err != nil {
		return opts, fmt.Errorf("error creating OCM client: %v", err)
	}
	defer ocmClient.Close()

	clusterId, err := ocm.GetClusterId(ocmClient, cluster)
	if err != nil {
		return opts, fmt.Errorf("error getting cluster ID from OCM: %v", err)
	}

	// Determine the storage directory
	storageDir, err := f.statStorageDir()
	if err != nil {
		return opts, fmt.Errorf("error locating storage directory: %v", err)
	}

	// Create the mount directory for this cluster
	mount := filepath.Join(storageDir, clusterId)
	log.Debugf("ensuring filepath exists for cluster: %s", mount)
	err = f.afs.MkdirAll(mount, os.ModePerm)
	if err != nil {
		f.criticalError = true
		return opts, fmt.Errorf("error creating persistent histories directory: %v", err)
	}

	opts.AddVolumeMount(engine.VolumeMount{
		Source:       mount,
		Destination:  destDir,
		MountOptions: "rw",
	})

	opts.AddEnv(engine.EnvVar{
		Key:   "HISTFILE",
		Value: destDir + "/" + histFile,
	})

	return opts, nil
}

func (f *Feature) HandleError(err error) {
	if f.userHasConfig {
		log.Warnf("Error initializing persistent histories functionality: %v", err)
	}
	log.Debugf("Error initializing persistent histories functionality: %v", err)
}

// statStorageDir checks for storage directory locations in the following order:
// absolute path -> $HOME/(path)
// If the directory doesn't exist, returns the path anyway (will be created later)
func (f *Feature) statStorageDir() (string, error) {
	dirpath := f.config.StorageDir
	if dirpath == "" {
		return "", fmt.Errorf("no storage directory provided")
	}
	errorPaths := []string{}

	// Check absolute path first
	fileinfo, err := f.afs.Stat(dirpath)
	if err == nil {
		log.Debugf("using %s for persistent histories storage dir", dirpath)
		if !fileinfo.IsDir() {
			return "", fmt.Errorf("persistent history storage path is not a directory")
		}
		return dirpath, nil
	}
	errorPaths = append(errorPaths, dirpath)

	// Try $HOME-relative path
	path := filepath.Join(os.Getenv("HOME"), dirpath)
	fileinfo, err = f.afs.Stat(path)
	if err == nil {
		log.Debugf("using %s for persistent histories storage dir", path)
		if !fileinfo.IsDir() {
			return "", fmt.Errorf("persistent history storage path is not a directory")
		}
		return path, nil
	}
	errorPaths = append(errorPaths, path)

	return "", fmt.Errorf("could not find %s in any of: %s", dirpath, errorPaths)
}

func init() {
	f := Feature{
		afs: &afero.Afero{Fs: afero.NewOsFs()},
	}
	features.Register("persistent-histories", &f)
}
