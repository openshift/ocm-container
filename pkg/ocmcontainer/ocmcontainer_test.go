package ocmcontainer

import (
	"reflect"
	"testing"
)

func TestTTY(t *testing.T) {
	testCases := []struct {
		name     string
		input1   bool
		input2   bool
		expected []string
	}{
		{"Pseudo-tty and interactive specified", true, true, []string{"--tty", "--interactive"}},
		{"Neither pseudo-tty nor interactive specified", false, false, nil},
		{"Only pseudo-tty specified", true, false, []string{"--tty"}},
		{"Only interactive specified", false, true, nil},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tty(tc.input1, tc.input2)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected '%s', but got '%s'", tc.expected, result)
			}
		})
	}
}

func TestParseArgs(t *testing.T) {
	testCases := []struct {
		name      string
		input     []string
		expected1 string
		expected2 string
		// TODO: add errExpected
	}{
		{"One argument", []string{"arg1"}, "arg1", ""},
		{"Multiple arguments with no separator", []string{"arg1", "arg2", "arg3"}, "", ""},
		{"Three arguments with separator", []string{"arg1", "--", "arg2"}, "arg1", "arg2"},
		{"More than three arguments with separator", []string{"arg1", "--", "arg2", "arg3"}, "arg1", "arg2 arg3"},
		{"No arguments", []string{}, "", ""},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result1, result2, _ := parseArgs(tc.input)
			if result1 != tc.expected1 {
				t.Errorf("Expected '%s', but got '%s'", tc.expected1, result1)
			}
			if result2 != tc.expected2 {
				t.Errorf("Expected '%s', but got '%s'", tc.expected1, result1)
			}
		})
	}
}

// TODO: Figure out how to test parseFlags()
