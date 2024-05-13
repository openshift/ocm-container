package engine

import (
	"reflect"
	"testing"
)

func TestTtyToString(t *testing.T) {
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
			result := ttyToString(tc.input1, tc.input2)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected '%s', but got '%s'", tc.expected, result)
			}
		})
	}
}

func TestEnvsToString(t *testing.T) {
	testCases := []struct {
		name     string
		input    map[string]string
		expected []string
	}{
		{"Several fields specified", map[string]string{"key1": "val1", "key2": "val2"}, []string{"--env", "key1=val1", "--env", "key2=val2"}},
		{"One field specified", map[string]string{"key1": "val1"}, []string{"--env", "key1=val1"}},
		{"Mix of key and key/value field specified", map[string]string{"key1": "val1", "key2": ""}, []string{"--env", "key1=val1", "--env", "key2"}},
		{"Only the key is specified", map[string]string{"key1": ""}, []string{"--env", "key1"}},
		{"No fields specified", nil, nil},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := envsToString(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected '%s', but got '%s'", tc.expected, result)
			}
		})
	}
}

func TestParseRefToArgs(t *testing.T) {
	testCases := []struct {
		name      string
		container ContainerRef
		expected  string
	}{
		{
			name:      "Tests Publish All",
			container: ContainerRef{PublishAll: true},
			expected:  "--publish-all",
		},
	}

	// run once for special empty arg string case
	t.Run("Tests empty containerRef should return empty arg slice", func(t *testing.T) {
		args, err := parseRefToArgs(ContainerRef{})
		if err != nil {
			t.Errorf("Unexpected Error: %v", err)
		}
		if len(args) > 0 {
			t.Errorf("Expected empty arg slice, got len %d :: %v", len(args), args)
		}
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			args, err := parseRefToArgs(tc.container)
			if err != nil {
				t.Errorf("Expected no error but got %v", err)
			}
			for _, arg := range args {
				if arg == tc.expected {
					return
				}
			}
			t.Errorf("%s not found in args: %v", tc.expected, args)
		})
	}
}
