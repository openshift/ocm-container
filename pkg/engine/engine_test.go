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
		{
			"Several fields specified unsorted",
			map[string]string{"z_key1": "val1", "bkey3": "val3", "a_key2": "val2"},
			[]string{"--env", "a_key2=val2", "--env", "bkey3=val3", "--env", "z_key1=val1"},
		},
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
		expected  []string
	}{
		{
			name:      "Tests Publish All",
			container: ContainerRef{PublishAll: true},
			expected:  []string{"--publish-all"},
		},
		{
			name:      "Tests LocalPorts single",
			container: ContainerRef{LocalPorts: map[string]int{"console": 9999}},
			expected:  []string{"--publish=127.0.0.1::9999"},
		},
		{
			name:      "Tests LocalPorts multiple",
			container: ContainerRef{LocalPorts: map[string]int{"console": 9999, "promlens": 8080}},
			expected:  []string{"--publish=127.0.0.1::9999", "--publish=127.0.0.1::8080"},
		},
		{
			name:      "Tests privileged",
			container: ContainerRef{Privileged: true},
			expected:  []string{"--privileged"},
		},
		{
			name:      "Tests Remove after Exit",
			container: ContainerRef{RemoveAfterExit: true},
			expected:  []string{"--rm"},
		},
		{
			name:      "Tests bestEffortArgs",
			container: ContainerRef{BestEffortArgs: []string{"--abcd", "-v ~/path/in/filesystem:/root/folder"}},
			expected:  []string{"--abcd", "-v ~/path/in/filesystem:/root/folder"},
		},
		{
			name:      "Tests entrypoint",
			container: ContainerRef{Entrypoint: "abcdefg"},
			expected:  []string{"--entrypoint=abcdefg"},
		},
		{
			name:      "Tests tty",
			container: ContainerRef{Tty: true, Interactive: true},
			expected:  []string{"--interactive", "--tty"},
		},
		{
			name:      "Tests image fqdn",
			container: ContainerRef{Image: ContainerImage{Name: "imagename", Tag: "test"}},
			expected:  []string{"imagename:test"},
		},
		{
			name:      "Tests image fqdn with reponame",
			container: ContainerRef{Image: ContainerImage{Name: "imagename", Tag: "test", Repository: "openshift"}},
			expected:  []string{"openshift/imagename:test"},
		},
		{
			name:      "Tests image fqdn with reponame and registry",
			container: ContainerRef{Image: ContainerImage{Name: "imagename", Tag: "test", Repository: "openshift", Registry: "quay.io"}},
			expected:  []string{"quay.io/openshift/imagename:test"},
		},
		{
			name:      "Tests command",
			container: ContainerRef{Command: "oc do something here"},
			expected:  []string{"oc do something here"},
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

	// run once for special empty local port map case
	t.Run("Tests empty LocalPorts map returns no args", func(t *testing.T) {
		args, err := parseRefToArgs(ContainerRef{LocalPorts: map[string]int{}})
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

			// create the map to determine if all have been found
			expectedFound := map[string]bool{}
			for _, flag := range tc.expected {
				expectedFound[flag] = false
			}

			// loop through all args and mark as found
			for _, arg := range args {
				for _, expected := range tc.expected {
					if arg == expected {
						expectedFound[expected] = true
					}
				}
			}

			// loop through all of the expected to be found args and if
			// not all are true then return an error
			for _, found := range expectedFound {
				if !found {
					t.Errorf("%s not found in args: %v", tc.expected, args)
					return
				}
			}
		})
	}

	// Validate that the command comes last
	t.Run("Tests the command always comes last", func(t *testing.T) {
		container := ContainerRef{
			Command:         "my command to do something",
			RemoveAfterExit: true,
			Privileged:      true,
			Image: ContainerImage{
				Name: "test",
				Tag:  "latest",
			},
		}

		args, err := parseRefToArgs(container)
		if err != nil {
			t.Errorf("Unexpected Error: %v", err)
		}

		last := len(args) - 1
		if args[last] != container.Command {
			t.Errorf("Expected last argument to be \"%s\", but got: %s", container.Command, args[last])
		}
	})
}

func TestPullPolicyToString(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Always", "always", "--pull=always"},
		{"Missing", "missing", "--pull=missing"},
		{"Never", "never", "--pull=never"},
		{"Newer", "newer", "--pull=newer"},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := pullPolicyToString(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', but got '%s'", tc.expected, result)
			}
		})
	}
}
