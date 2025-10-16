package ocmcontainer

import (
	"slices"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// This package creates a "standard" set of ocm-container environment variables
// This is not intended to add any customization or logic - just return those envs that
// every container should have set

var skippedKeys = []string{
	"env",
	"environment",
	"volume",
	"volumes",
	"volumemounts",
	"vols",
}

// ocmContainerEnvs returns a map of environment variables for a standard ocm-container env
func ocmContainerEnvs() map[string]string {

	e := make(map[string]string)

	// Setting the strings to empty will pass them in
	// in the "-e USER" from the environment format

	// standard env vars specified as nil strings will be passed to the engine
	// in as "-e VAR" using the value from os.Environ() to the syscall.Exec() call

	e["USER"] = "" // THIS IS NOT A NIL VALUE INSIDE THE CONTAINER; SEE NOTE ABOVE

	// TODO: These should go in the envs.go, and perhaps
	// be a range over the viper.AllKeys() cross-referenced with
	// cmd.ManagedFields (configure?)

	// Parse all the viper key/values as OCMC_ envs inside the container
	for _, k := range viper.AllKeys() {

		var s string
		s = strings.ToUpper("OCMC_" + k)
		s = strings.ReplaceAll(s, "-", "_")
		i := viper.Get(k)
		log.Debugf("Parsing env %s :: %s", k, i)

		if slices.Contains(skippedKeys, k) {
			continue
		}
		switch v := i.(type) {
		case bool:
			e[s] = strconv.FormatBool(v)
		default:
			e[s] = v.(string)
		}
	}

	// Handle some deprecations
	e["INITIAL_CLUSTER_LOGIN"] = e["OCMC_CLUSTER_ID"]

	return e
}
