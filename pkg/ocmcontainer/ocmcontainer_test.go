package ocmcontainer

import (
	"testing"
)

func TestParseArgs(t *testing.T) {
	testCases := []struct {
		name      string
		input     []string
		expected1 string
		expected2 string
		// TODO: add errExpected
	}{
		{"One argument", []string{"arg1"}, "arg1", ""},
		{"Multiple arguments with no separator", []string{"arg1", "arg2", "arg3"}, "arg1", "arg2 arg3"},
		{"Three arguments with separator", []string{"arg1", "--", "arg2"}, "arg1", "arg2"},
		{"More than three arguments with separator", []string{"arg1", "--", "arg2", "arg3"}, "arg1", "arg2 arg3"},
		{"No arguments", []string{}, "", ""},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result1, result2, _ := parseArgs(tc.input)
			if result1 != tc.expected1 {
				t.Errorf("Expected first arg '%s', but got '%s'", tc.expected1, result1)
			}
			if result2 != tc.expected2 {
				t.Errorf("Expected second arg '%s', but got '%s'", tc.expected2, result2)
			}
		})
	}
}

// TODO: Figure out how to test parseFlags()
