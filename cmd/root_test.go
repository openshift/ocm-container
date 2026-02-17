package cmd

import (
	"reflect"
	"testing"
)

func TestSplitArgs(t *testing.T) {
	testCases := []struct {
		name          string
		input         []string
		expectedCobra []string
		expectedExec  []string
	}{
		{
			name:          "Empty args",
			input:         []string{},
			expectedCobra: nil,
			expectedExec:  nil,
		},
		{
			name:          "Only program name",
			input:         []string{"ocm-container"},
			expectedCobra: nil,
			expectedExec:  nil,
		},
		{
			name:          "Flags only",
			input:         []string{"ocm-container", "--cluster-id", "test-cluster"},
			expectedCobra: []string{"--cluster-id", "test-cluster"},
			expectedExec:  nil,
		},
		{
			name:          "Double dash with command",
			input:         []string{"ocm-container", "--", "echo", "hello"},
			expectedCobra: []string{},
			expectedExec:  []string{"echo", "hello"},
		},
		{
			name:          "Flags and double dash with command",
			input:         []string{"ocm-container", "--cluster-id", "test-cluster", "--", "oc", "get", "pods"},
			expectedCobra: []string{"--cluster-id", "test-cluster"},
			expectedExec:  []string{"oc", "get", "pods"},
		},
		{
			name:          "Double dash with single command",
			input:         []string{"ocm-container", "--", "bash"},
			expectedCobra: []string{},
			expectedExec:  []string{"bash"},
		},
		{
			name:          "Multiple flags before double dash",
			input:         []string{"ocm-container", "--verbose", "--cluster-id", "cluster1", "--", "command"},
			expectedCobra: []string{"--verbose", "--cluster-id", "cluster1"},
			expectedExec:  []string{"command"},
		},
		{
			name:          "Double dash at the end with no command",
			input:         []string{"ocm-container", "--cluster-id", "test", "--"},
			expectedCobra: []string{"--cluster-id", "test"},
			expectedExec:  []string{},
		},
		{
			name:          "No double dash",
			input:         []string{"ocm-container", "--cluster-id", "test", "--verbose"},
			expectedCobra: []string{"--cluster-id", "test", "--verbose"},
			expectedExec:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cobraArgs, execArgs := splitArgs(tc.input)

			if !reflect.DeepEqual(cobraArgs, tc.expectedCobra) {
				t.Errorf("Cobra args mismatch:\n  got:  %v\n  want: %v", cobraArgs, tc.expectedCobra)
			}

			if !reflect.DeepEqual(execArgs, tc.expectedExec) {
				t.Errorf("Exec args mismatch:\n  got:  %v\n  want: %v", execArgs, tc.expectedExec)
			}
		})
	}
}
