package ocmcontainer

import (
	"testing"

	"github.com/spf13/viper"
)

// TODO: Figure out how to test parseFlags()
func TestParseArgs(t *testing.T) {
	testCases := []struct {
		name        string
		input       []string
		cluster     string
		expected1   string
		expected2   string
		errExpected error
	}{
		{"One argument nil cluster", []string{"arg1"}, "", "arg1", "", nil},
		{"One argument with cluster", []string{"arg1"}, "clusterName", "clusterName", "arg1", nil},
		{"Multiple arguments nil cluster", []string{"arg1", "arg2", "arg3"}, "", "arg1", "arg2 arg3", nil},
		{"Multiple arguments with cluster", []string{"arg1", "arg2", "arg3"}, "clusterName", "clusterName", "arg1 arg2 arg3", nil},
		{"No arguments nil cluster", []string{}, "", "", "", nil},
		{"No arguments with cluster", []string{}, "clusterName", "clusterName", "", nil},
		{"Multiple arguments with cluster with a dash", []string{"-", "arg2_arg3"}, "clusterName", "", "", errClusterAndDashArgs},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result1, result2, err := parseArgs(tc.input, tc.cluster)
			if result1 != tc.expected1 {
				t.Errorf("Expected first arg '%s', but got '%s'", tc.expected1, result1)
			}
			if result2 != tc.expected2 {
				t.Errorf("Expected second arg '%s', but got '%s'", tc.expected2, result2)
			}

			if err != nil && err != tc.errExpected {
				t.Errorf("Expected error '%v', but got '%v'", tc.errExpected, err)
			}
		})
	}
}

// Features which are disabled with --no-something=true
// should return false when looking up enabled(something)
func TestFeatureEnabled(t *testing.T) {
	viper.Set("no-foo", false)
	viper.Set("no-bar", true)

	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Feature enabled", "foo", true},
		{"Feature disabled", "bar", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := featureEnabled(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%t', but got '%t'", tc.expected, result)
			}
		})
	}
}

func TestLookUpNegativeName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Convert feature name to negative", "foo", "no-foo"},
		{"Convert feature with dash to negative", "bar-baz", "no-bar-baz"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := lookUpNegativeName(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', but got '%s'", tc.expected, result)
			}
		})
	}
}
