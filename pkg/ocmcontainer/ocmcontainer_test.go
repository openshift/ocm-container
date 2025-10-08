package ocmcontainer

import (
	"errors"
	"testing"

	"github.com/openshift/ocm-container/pkg/engine"
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

			if err != nil && !errors.Is(err, tc.errExpected) {
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

func TestParseMountString(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      engine.VolumeMount
		expectedError bool
	}{
		{
			name:  "basic mount with source and destination",
			input: "/path/to/local:/path/in/container",
			expected: engine.VolumeMount{
				Source:       "/path/to/local",
				Destination:  "/path/in/container",
				MountOptions: "",
			},
			expectedError: false,
		},
		{
			name:  "mount with read-only option",
			input: "/path/to/local:/path/in/container:ro",
			expected: engine.VolumeMount{
				Source:       "/path/to/local",
				Destination:  "/path/in/container",
				MountOptions: "ro",
			},
			expectedError: false,
		},
		{
			name:  "mount with read-write option",
			input: "/path/to/local:/path/in/container:rw",
			expected: engine.VolumeMount{
				Source:       "/path/to/local",
				Destination:  "/path/in/container",
				MountOptions: "rw",
			},
			expectedError: false,
		},
		{
			name:  "mount with complex options",
			input: "/path/to/local:/path/in/container:ro,z",
			expected: engine.VolumeMount{
				Source:       "/path/to/local",
				Destination:  "/path/in/container",
				MountOptions: "ro,z",
			},
			expectedError: false,
		},
		{
			name:  "mount with relative source path",
			input: "./relative/path:/path/in/container",
			expected: engine.VolumeMount{
				Source:       "./relative/path",
				Destination:  "/path/in/container",
				MountOptions: "",
			},
			expectedError: false,
		},
		{
			name:  "mount with home directory expansion",
			input: "~/config:/root/.config",
			expected: engine.VolumeMount{
				Source:       "~/config",
				Destination:  "/root/.config",
				MountOptions: "",
			},
			expectedError: false,
		},
		{
			name:          "empty string",
			input:         "",
			expected:      engine.VolumeMount{},
			expectedError: true,
		},
		{
			name:          "only source path",
			input:         "/path/to/local",
			expected:      engine.VolumeMount{},
			expectedError: true,
		},
		{
			name:          "missing source path",
			input:         ":/path/in/container",
			expected:      engine.VolumeMount{},
			expectedError: true,
		},
		{
			name:          "missing destination path",
			input:         "/path/to/local:",
			expected:      engine.VolumeMount{},
			expectedError: true,
		},
		{
			name:          "too many colons",
			input:         "/path/to/local:/path/in/container:ro:extra",
			expected:      engine.VolumeMount{},
			expectedError: true,
		},
		{
			name:          "missing both paths",
			input:         ":",
			expected:      engine.VolumeMount{},
			expectedError: true,
		},
		{
			name:          "empty source with options",
			input:         ":/path/in/container:ro",
			expected:      engine.VolumeMount{},
			expectedError: true,
		},
		{
			name:          "empty destination with options",
			input:         "/path/to/local::ro",
			expected:      engine.VolumeMount{},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseMountString(tt.input)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Source != tt.expected.Source {
				t.Errorf("Source mismatch: got %q, want %q", result.Source, tt.expected.Source)
			}

			if result.Destination != tt.expected.Destination {
				t.Errorf("Destination mismatch: got %q, want %q", result.Destination, tt.expected.Destination)
			}

			if result.MountOptions != tt.expected.MountOptions {
				t.Errorf("MountOptions mismatch: got %q, want %q", result.MountOptions, tt.expected.MountOptions)
			}
		})
	}
}
