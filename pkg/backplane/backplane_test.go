package backplane

import (
	"errors"
	"io/fs"
	"os"
	"testing"

	"github.com/openshift/ocm-container/pkg/engine"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name              string
		homeInput         string
		backplaneEnvInput string
		expected          *Config
		errExpected       error
	}{
		{
			name:              "BACKPLANE_CONFIG environment variable is set",
			homeInput:         "/home/user",
			backplaneEnvInput: "/home/user/.config/backplane/prod-config.json",
			expected: &Config{
				Env: map[string]string{
					"BACKPLANE_CONFIG": backplaneConfigDest,
				},
				Mounts: []engine.VolumeMount{
					{
						Source:       "/home/user/.config/backplane/prod-config.json",
						Destination:  backplaneConfigDest,
						MountOptions: backplaneConfigMountOpts,
					},
				},
			},
			errExpected: nil,
		},
		{
			name:              "BACKPLANE_CONFIG environment variable is not set",
			homeInput:         "/home/user",
			backplaneEnvInput: "",
			expected: &Config{
				Env: map[string]string{
					"BACKPLANE_CONFIG": backplaneConfigDest,
				},
				Mounts: []engine.VolumeMount{
					{
						Source:       "/home/user/.config/backplane/config.json",
						Destination:  backplaneConfigDest,
						MountOptions: backplaneConfigMountOpts,
					},
				},
			},
			errExpected: nil,
		},
		{
			name:              "BACKPLANE_CONFIG environment variable is set but file does not exist",
			homeInput:         "/home/user",
			backplaneEnvInput: "/home/user/.config/backplane/nonexistent-config.json",
			expected:          &Config{},
			errExpected: &fs.PathError{
				Op:   "stat",
				Path: "/home/user/.config/backplane/nonexistent-config.json",
				Err:  errors.New("no such file or directory"),
			},
		},
		{
			name:              "BACKPLANE_CONFIG environment variable and HOME do not match (should not matter)",
			homeInput:         "/home/some-other-user",
			backplaneEnvInput: "/home/user/.config/backplane/prod-config.json",
			expected: &Config{
				Env: map[string]string{
					"BACKPLANE_CONFIG": backplaneConfigDest,
				},
				Mounts: []engine.VolumeMount{
					{
						Source:       "/home/user/.config/backplane/prod-config.json",
						Destination:  backplaneConfigDest,
						MountOptions: backplaneConfigMountOpts,
					},
				},
			},
			errExpected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stat := osStat
			defer func() { osStat = stat }()

			osStat = func(name string) (os.FileInfo, error) {
				if tc.name == "BACKPLANE_CONFIG environment variable is set but file does not exist" {
					return nil, errors.New("someError")
				}
				return nil, nil
			}

			err := os.Setenv("BACKPLANE_CONFIG", tc.backplaneEnvInput)
			if err != nil {
				t.Errorf("Test Setup Failed: error setting environment variable: %v", err)
			}

			config, err := New(tc.homeInput)

			// Assert that the error is not nil if we expect an error
			if err == nil && tc.errExpected != nil {
				t.Errorf("Expected: %v (%T), got %v (%T)", tc.errExpected, tc.errExpected, err, err)
			}

			// Assert that the BACKPLANE_CONFIG environment variable is set correctly
			if config.Env["BACKPLANE_CONFIG"] != tc.expected.Env["BACKPLANE_CONFIG"] {
				t.Errorf("Expected BACKPLANE_CONFIG to be %v, got %v", tc.expected.Env["BACKPLANE_CONFIG"], config.Env["BACKPLANE_CONFIG"])
			}

			if tc.errExpected == nil {
				if len(config.Mounts) != len(tc.expected.Mounts) {
					t.Errorf("Expected one mount, got %v", len(config.Mounts))
				} else {
					mount := config.Mounts[0]
					if mount.Source != tc.expected.Mounts[0].Source || mount.Destination != tc.expected.Mounts[0].Destination || mount.MountOptions != tc.expected.Mounts[0].MountOptions {
						t.Errorf("Expected mount values %v, got %v", tc.expected.Mounts, mount)
					}
				}
			}
		})
	}
}
