package ocmcontainer

import (
	"testing"
)

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
		{"Multiple arguments with cluster with an underscore", []string{"_", "arg2_arg3"}, "clusterName", "", "", errClusterAndUnderscoreArgs},
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

// TODO: Figure out how to test parseFlags()
