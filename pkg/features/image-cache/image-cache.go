package imagecache

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/ocm-container/pkg/engine"
	"github.com/openshift/ocm-container/pkg/features"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

const (
	FeatureFlagName = "no-image-cache"
	FlagHelpMessage = "Disable persistent container image caching"

	destDir           = "/var/lib/containers/storage/"
	defaultStorageDir = ".config/ocm-container/images"
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
		log.Debugf("image-cache disabled via config")
		return false
	}
	if viper.IsSet(FeatureFlagName) {
		log.Debugf("image-cache disabled via flag")
		return false
	}
	return true
}

func (f *Feature) ExitOnError() bool {
	return f.criticalError
}

func (f *Feature) Configure() error {
	cfg := newConfigWithDefaults()

	if !viper.IsSet("features.image_cache") {
		f.config = cfg
		return nil
	}

	f.userHasConfig = true
	err := viper.UnmarshalKey("features.image_cache", &cfg)
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

	// Determine the storage directory
	storageDir, err := f.statStorageDir()
	if err != nil {
		return opts, fmt.Errorf("error locating storage directory: %v", err)
	}

	opts.AddVolumeMount(engine.VolumeMount{
		Source:       storageDir,
		Destination:  destDir,
		MountOptions: "rw",
	})

	return opts, nil
}

func (f *Feature) HandleError(err error) {
	if f.userHasConfig {
		log.Warnf("Error initializing image cache functionality: %v", err)
	}
	log.Debugf("Error initializing image cache functionality: %v", err)
}

// statStorageDir checks for storage directory locations in the following order:
// absolute path -> $HOME/(path)
// Returns an error if the directory doesn't exist in any of these locations
func (f *Feature) statStorageDir() (string, error) {
	dirpath := f.config.StorageDir
	if dirpath == "" {
		return "", fmt.Errorf("no storage directory provided")
	}

	// Check absolute path first
	fileinfo, err := f.afs.Stat(dirpath)
	if err == nil {
		log.Debugf("using %s for image cache storage dir", dirpath)
		if !fileinfo.IsDir() {
			f.criticalError = true
			return "", fmt.Errorf("image cache storage path is not a directory")
		}
		return dirpath, nil
	}

	// Try $HOME-relative path
	path := filepath.Join(os.Getenv("HOME"), dirpath)
	fileinfo, err = f.afs.Stat(path)
	if err == nil {
		log.Debugf("using %s for image cache storage dir", path)
		if !fileinfo.IsDir() {
			f.criticalError = true
			return "", fmt.Errorf("image cache storage path is not a directory")
		}
		return path, nil
	}

	// If neither exists, return an error
	errorPaths := []string{dirpath, path}
	return "", fmt.Errorf("could not find %s in any of: %v", dirpath, errorPaths)
}

func init() {
	f := Feature{
		afs: &afero.Afero{Fs: afero.NewOsFs()},
	}
	features.Register("image-cache", &f)
}
