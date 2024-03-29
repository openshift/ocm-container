/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type legacyField struct {
	name           string
	sensitive      bool
	deprecated     bool
	deprecationMsg string
}

//lint:file-ignore U1000 field-type, managedFields, et all are for future use
type fieldType int

const (
	required fieldType = iota
	optional
	legacy
)

const (
	sshDeprecationMsg                = "SSH multiplexing and Socket mounting is no longer needed or supported. Please remove the 'DISABLE_SSH_MULTIPLEXING' and 'SSH_AUTH_SOCK' fields from your configuration."
	backplaneConfigDirDeprecationMsg = "The 'BACKPLANE_CONFIG_DIR' field is deprecated and will be removed in a future version. Please remove it from your configuration.  You may specify an alternate backplane config file with 'BACKPLANE_CONFIG'."
	ocmUrlDeprecationMsg             = "The 'OCM_URL' field is deprecated and will be removed in a future version. Please remove it from your configuration."
	ocmUserDeprecationMsg            = "The 'OCM_USER' field is deprecated and will be removed in a future version. Please remove it from your configuration."
	cliDeprecationMsg                = "The 'CLI' field is deprecated and will be removed in a future version. Please remove it from your configuration."
)

type managedField struct {
	name      string
	sensitive bool
	fieldType fieldType
}

var (
	sensitiveDataFields = []string{
		"offline_access_token",
		"OFFLINE_ACCESS_TOKEN",
	}

	ManagedFields = []managedField{
		{"backplane_config_dir", false, required},
		{"ca_source_anchors", false, optional},
		{"engine", false, required},
		{"jira_token", true, optional},
		{"ocm_url", false, optional},
		{"offline_access_token", true, required},
		{"pagerduty_token", true, optional},
	}

	legacyCfgFile = os.Getenv("HOME") + "/.config/ocm-container/env.source"
	legacyFields  = map[string]legacyField{
		"BACKPLANE_CONFIG_DIR":         {deprecated: true, deprecationMsg: backplaneConfigDirDeprecationMsg, name: "BACKPLANE_CONFIG_DIR"},
		"CA_SOURCE_ANCHORS":            {name: "CA_SOURCE_ANCHORS"},
		"CLI":                          {deprecated: true, deprecationMsg: cliDeprecationMsg, name: "CLI"},
		"CONTAINER_SUBSYS":             {name: "engine"},
		"DISABLE_SSH_MULTIPLEXING":     {deprecated: true, deprecationMsg: sshDeprecationMsg, name: "DISABLE_SSH_MULTIPLEXING"},
		"OCM_URL":                      {deprecated: true, deprecationMsg: ocmUrlDeprecationMsg, name: "OCM_URL"},
		"OCM_USER":                     {deprecated: true, deprecationMsg: ocmUserDeprecationMsg, name: "OCM_USER"},
		"OFFLINE_ACCESS_TOKEN":         {sensitive: true, name: "OFFLINE_ACCESS_TOKEN"},
		"OPS_UTILS_DIR":                {name: "OPS_UTILS_DIR"},
		"OPS_UTILS_DIR_RW":             {name: "OPS_UTILS_DIR_RW"},
		"PATH":                         {name: "PATH"},
		"PERSISTENT_CLUSTER_HISTORIES": {name: "PERSISTENT_CLUSTER_HISTORIES"},
		"PERSONALIZATION_FILE":         {name: "PERSONALIZATION_FILE"},
		"SCRATCH_DIR":                  {name: "SCRATCH_DIR"},
		"SSH_AUTH_SOCK":                {deprecated: true, deprecationMsg: sshDeprecationMsg, name: "SSH_AUTH_SOCK"},
	}
)

// configureCmd sets or gets the configuration of the application
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "View or Set the configuration of the application",
	Long: `The configure command allows viewing or setting the
configuration of the program field by field, or optionally
through a command line TUI prompt.`,
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific field (or all fields if none provided) of the application configuration",
	// At some point this could be expanded to retrieve a specific field
	// ValidArgs: validConfigArgs,
	Args: cobra.MatchAll(cobra.MaximumNArgs(2), cobra.ArbitraryArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		showSensitiveValues, err := cmd.Flags().GetBool("show-sensitive-values")
		if err != nil {
			return err
		}

		var k string
		if len(args) == 0 {
			k = "all"
		} else {
			k = args[0]
		}

		err = getValue(k, viper.GetViper(), showSensitiveValues)
		if err != nil {
			return err
		}

		return nil
	},
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set an individual key:value field of the application configuration",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			return err
		}

		showSensitiveValues, err := cmd.Flags().GetBool("show-sensitive-values")
		if err != nil {
			return err
		}

		setValue(args[0], args[1], showSensitiveValues, dryRun)
		if !dryRun {
			err = viper.WriteConfig()
			if err != nil {
				return err
			}
		}

		fmt.Printf("Configuration written to %s\n", cfgFile)

		return nil
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Launch a TUI prompt to set the configuration of the application",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		var existingConfig bool = true
		var legacyData map[string]any

		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			return err
		}

		showSensitiveValues, err := cmd.Flags().GetBool("show-sensitive-values")
		if err != nil {
			return err
		}

		assumeYes, err := cmd.Flags().GetBool("assume-yes")
		if err != nil {
			return err
		}

		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			existingConfig = false
		}

		if _, err := os.Stat(legacyCfgFile); !os.IsNotExist(err) {
			legacyData, err = readLegacyConfig()
			if err != nil {
				// If we can't read the legacy config,
				// we'll just start fresh
				legacyData = nil
			}
		}

		var newConfig map[string]string
		newConfig, err = buildNewConfig(legacyData, existingConfig, showSensitiveValues, assumeYes)
		if err != nil {
			return err
		}

		viper.SetConfigFile(cfgFile)

		for k, v := range newConfig {
			setValue(k, v, showSensitiveValues, dryRun)
		}

		if !dryRun {
			err = viper.WriteConfig()
			if err != nil {
				return err
			}
		}

		fmt.Printf("Configuration written to %s\n", cfgFile)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)

	configureCmd.AddCommand(getCmd)
	configureCmd.AddCommand(setCmd)
	configureCmd.AddCommand(initCmd)

}

func init() {
	// Get
	getCmd.Flags().Bool("show-sensitive-values", false, "prints potentially sensitive values (tokens, etc) rather than redacting them")

	// Set
	setCmd.Flags().Bool("dry-run", false, "parses arguments and environment and prints the data that would be written, but does not write it.")
	setCmd.Flags().Bool("show-sensitive-values", false, "prints potentially sensitive values (tokens, etc) rather than redacting them")

	// Init
	initCmd.Flags().Bool("dry-run", false, "parses arguments and environment and prints the data that would be written, but does not write it.")
	initCmd.Flags().BoolP("assume-yes", "y", false, "Assume 'yes' as the answer to any prompts and run non-interactively (best effort)")
	initCmd.Flags().Bool("show-sensitive-values", false, "prints potentially sensitive values (tokens, etc) rather than redacting them")
}

func setValue(k, v string, showSensitiveValues, dryRun bool) {
	if dryRun {
		nopeRope := viper.New()
		nopeRope.Set(k, v)
		fmt.Println("dry-run - would have set:")
		getValue(k, nopeRope, showSensitiveValues)
		getValue("all", viper.GetViper(), showSensitiveValues)
		return
	}

	viper.Set(k, v)
}

func getValue(k string, viper *viper.Viper, showSensitiveValues bool) error {
	var s strings.Builder

	switch k {
	case "all":
		for k, v := range viper.AllSettings() {
			if slices.Contains(sensitiveDataFields, k) && !showSensitiveValues {
				s.WriteString(fmt.Sprintf("%s: %v\n", k, "REDACTED"))
			} else {
				s.WriteString(fmt.Sprintf("%s: %v\n", k, v))
			}
		}
		fmt.Print(s.String())

	default:
		v := viper.Get(k)
		if slices.Contains(sensitiveDataFields, k) && !showSensitiveValues {
			fmt.Printf("%s: %v\n", k, "REDACTED")
		} else {
			fmt.Printf("%s: %v\n", k, v)
		}
	}

	return nil
}

func readLegacyConfig() (map[string]any, error) {
	data := make(map[string]any)

	var potentialParseErrors []string

	f, err := os.Open(legacyCfgFile)
	//lint:ignore SA5001 setting defer before the error check,
	// because the function will return immediately with the error
	defer f.Close()
	if err != nil {
		return data, err
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "=") && !strings.HasPrefix(line, "#") {
			line = os.ExpandEnv(line)
			line = strings.TrimLeft(line, "export ")
			parts := strings.SplitN(line, "=", 2)
			_, ok := legacyFields[parts[0]]
			if !ok || parts[1] == "" {
				potentialParseErrors = append(potentialParseErrors, line)
			} else {
				data[parts[0]] = parts[1]
			}
		}
	}

	if len(potentialParseErrors) > 0 {
		data["potentialParseErrors"] = potentialParseErrors
	}

	return data, nil
}

func promptUserToWriteConfig(existingConfig bool) (bool, error) {
	var configPrompt *huh.Confirm
	var writeConfig bool

	if existingConfig {
		configPrompt = huh.NewConfirm().
			Title("Configuration file already exists. Would you like to overwrite it?").
			Affirmative("yes").
			Negative("no").
			Value(&writeConfig)
	} else {
		configPrompt = huh.NewConfirm().
			Title("No configuration file found. Would you like to create one?").
			Affirmative("yes").
			Negative("no").
			Value(&writeConfig)
	}

	var err error
	var groups []*huh.Group

	groups = append(groups, huh.NewGroup(configPrompt))
	err = huh.NewForm(groups...).WithShowHelp(true).WithTheme(huh.ThemeBase()).WithProgramOptions(tea.WithAltScreen()).Run()
	if err != nil {
		return false, err
	}

	return writeConfig, nil
}

func promptUserToImportLegacyFields() (bool, error) {
	var legacyPrompt *huh.Confirm
	var importLegacyConfig bool

	legacyPrompt = huh.NewConfirm().
		Title(fmt.Sprintf("Try to import values from the legacy configuration file '%s'?", legacyCfgFile)).
		Affirmative("yes").
		Negative("no").
		Value(&importLegacyConfig)

	var groups []*huh.Group
	var err error

	groups = append(groups, huh.NewGroup(legacyPrompt))
	err = huh.NewForm(groups...).WithShowHelp(true).WithTheme(huh.ThemeBase()).WithProgramOptions(tea.WithAltScreen()).Run()

	if err != nil {
		return false, err
	}

	return importLegacyConfig, nil
}

func parseFieldsToValidAndScary(legacyData map[string]any) (map[string]any, map[string]any) {
	var validLegacyFields = make(map[string]any)
	var potentialParseErrors = make(map[string]any)

	for k, v := range legacyData {
		if k == "potentialParseErrors" {
			for _, e := range v.([]string) {
				parts := strings.SplitN(e, "=", 2)
				potentialParseErrors[parts[0]] = parts[1]
			}
		} else {
			validLegacyFields[k] = v.(string)
		}
	}

	return validLegacyFields, potentialParseErrors
}

func redactSensitiveFields(fields map[string]any) map[string]any {
	var redactedFields = make(map[string]any)

	for k, v := range fields {
		if legacyFields[k].sensitive {
			redactedFields[k] = "REDACTED"
		} else {
			redactedFields[k] = v
		}
	}
	return redactedFields
}

func convertFieldMapToOptions(fields map[string]any, showSensitiveValues bool) []huh.Option[string] {
	var s []string
	var redactedFields map[string]any

	if showSensitiveValues {
		redactedFields = fields
	} else {
		redactedFields = redactSensitiveFields(fields)
	}

	for k, v := range redactedFields {
		s = append(s, fmt.Sprintf("%s: %s", k, v))
	}

	sort.Strings(s)
	return huh.NewOptions(s...)
}

func promptUserToSelectFieldsToImport(legacyData map[string]any, showSensitiveValues bool) ([]string, error) {

	if legacyData == nil {
		return nil, nil
	}

	var fieldsToImport []string
	var scaryFieldsToImport []string

	validLegacyFields, potentialParseErrors := parseFieldsToValidAndScary(legacyData)

	opts := convertFieldMapToOptions(validLegacyFields, showSensitiveValues)
	titleString := fmt.Sprintf("Fields identified in legacy %s config file - select fields to import:", legacyCfgFile)
	fieldSelect := huh.NewMultiSelect[string]().
		Title(titleString).
		Options(opts...).
		Limit(len(opts)).
		Value(&fieldsToImport)

	scaryOpts := convertFieldMapToOptions(potentialParseErrors, showSensitiveValues)
	scaryFieldSelect := huh.NewMultiSelect[string]().
		Title("In addition, there may have been errors parsing these fields. Select any you would like to import:").
		Options(scaryOpts...).
		Limit(len(scaryOpts)).
		Value(&scaryFieldsToImport)

	var groups []*huh.Group
	var err error

	groups = append(groups, huh.NewGroup(fieldSelect, scaryFieldSelect))
	err = huh.NewForm(groups...).WithShowHelp(true).WithTheme(huh.ThemeBase()).WithProgramOptions(tea.WithAltScreen()).Run()
	if err != nil {
		return nil, err
	}

	var configSlice []string
	configSlice = append(configSlice, fieldsToImport...)
	configSlice = append(configSlice, scaryFieldsToImport...)

	return configSlice, nil
}

func promptUserToConfirmFields(slice []string) (bool, error) {
	var configPrompt *huh.Confirm
	var writeConfig bool

	var s strings.Builder
	s.WriteString("Importing these values to the new config; does this look OK?\n\n")
	for _, x := range slice {
		s.WriteString(fmt.Sprintf("%s\n", x))
	}

	configPrompt = huh.NewConfirm().
		Title(s.String()).
		Affirmative("yes").
		Negative("no").
		Value(&writeConfig)

	var groups []*huh.Group
	var err error

	groups = append(groups, huh.NewGroup(configPrompt))
	err = huh.NewForm(groups...).WithShowHelp(true).WithTheme(huh.ThemeBase()).WithProgramOptions(tea.WithAltScreen()).Run()

	if err != nil {
		return false, err
	}

	return writeConfig, nil
}

func promptUserToStartOver() (bool, error) {
	var startOverPrompt *huh.Confirm
	var startOver bool

	startOverPrompt = huh.NewConfirm().
		Title("Do you want to start over?").
		Affirmative("yes").
		Negative("no").
		Value(&startOver)

	var groups []*huh.Group
	var err error

	groups = append(groups, huh.NewGroup(startOverPrompt))
	err = huh.NewForm(groups...).WithShowHelp(true).WithTheme(huh.ThemeBase()).WithProgramOptions(tea.WithAltScreen()).Run()

	if err != nil {
		return false, err
	}

	// This lets us break out if the user doesn't want to start over;
	// and passes the huh.ErrUserAborted error up the stack
	if !startOver {
		return startOver, huh.ErrUserAborted
	}

	return startOver, nil
}

func buildNewConfig(legacyData map[string]any, existingConfig, showSensitiveValues, assumeYes bool) (map[string]string, error) {
	var err error

	var writeConfig bool
	var importLegacyConfig bool
	var configSlice []string

	// if assumeYes is true, auto-parse all the fields and
	// create the config with any valid fields, ignoring any
	// potentially problematic fields

	if assumeYes {
		validFields, _ := parseFieldsToValidAndScary(legacyData)

		configSlice := func(m map[string]any) []string {
			var s []string
			for k, v := range m {
				s = append(s, fmt.Sprintf("%s: %s", k, v))
			}
			sort.Strings(s)
			return s
		}(validFields)

		configMap := createConfigurationMap(legacyData, configSlice)
		return configMap, nil
	}

	// If assumeYes is not true, prompt the user to write a new config
	if !assumeYes {
		writeConfig, err = promptUserToWriteConfig(existingConfig)
		if err != nil {
			return nil, err
		}
	}

	// If the user doesn't want to write a new config, we're done
	if !writeConfig {
		return nil, huh.ErrUserAborted
	}

	// If there's legacy data, prompt to import
	if legacyData != nil && !assumeYes {
		importLegacyConfig, err = promptUserToImportLegacyFields()
		if err != nil {
			return nil, err
		}
	}

	// If the user doesn't want to import the legacy config, move to writing the new config
	if !importLegacyConfig {
		return nil, nil
	}

	// Prompt, and allow re-prompt until OK or Error
	for {
		// If the user does want to import the legacy data, prompt to select fields
		configSlice, err = promptUserToSelectFieldsToImport(legacyData, showSensitiveValues)
		if err != nil {
			return nil, err
		}

		// Let the user validate the fields to import
		writeConfig, err = promptUserToConfirmFields(configSlice)
		if err != nil {
			return nil, err
		}

		if !writeConfig {
			writeConfig, err = promptUserToStartOver()
		}
		if err != nil {
			return nil, err
		}

		// If the user is happy with the fields to import, break the loop
		if writeConfig {
			break
		}
	}

	configMap := createConfigurationMap(legacyData, configSlice)
	return configMap, nil
}

func createConfigurationMap(legacyData map[string]any, configSlice []string) map[string]string {
	var cm = make(map[string]string)

	validLegacyFields, potentialParseErrors := parseFieldsToValidAndScary(legacyData)

	// Split the confirmed string to get a key
	// select the value of the key from the legacy data
	// and add them to the new map
	for _, x := range configSlice {
		parts := strings.SplitN(x, ":", 2)

		var k string
		var v string

		// if the key is in the legacy fields map, use the name field from the value for that key
		if val, ok := legacyFields[parts[0]]; ok {
			k = val.name
		} else {
			k = parts[0]
		}

		// The value is the second part of the split from configSlice contains potentially
		// REDACTED data, so we need to look up the keys from configSlice in the legacyData
		// to get the original data.

		if val, ok := validLegacyFields[parts[0]]; ok {
			v = val.(string)
		} else if val, ok := potentialParseErrors[parts[0]]; ok {
			v = val.(string)
		} else {
			// Error finding the key in either map
			v = "parseError"
		}

		// Strip bash-y quotes
		v = strings.Trim(v, "\"")
		cm[k] = v
	}

	return cm
}
