package utils

import (
	"testing"
)

func TestIsRunningInOcmContainer(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{
			name:     "Running in ocm-container",
			envValue: "ocm-container",
			expected: true,
		},
		{
			name:     "Not in ocm-container - empty",
			envValue: "",
			expected: false,
		},
		{
			name:     "Not in ocm-container - different value",
			envValue: "some-other-container",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("IO_OPENSHIFT_MANAGED_NAME", tt.envValue)
			got := IsRunningInOcmContainer()
			if got != tt.expected {
				t.Errorf("IsRunningInOcmContainer() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetOcmContainerComponent(t *testing.T) {
	tests := []struct {
		name          string
		managedName   string
		componentName string
		expected      string
	}{
		{
			name:          "Micro variant",
			managedName:   "ocm-container",
			componentName: "micro",
			expected:      "micro",
		},
		{
			name:          "Minimal variant",
			managedName:   "ocm-container",
			componentName: "minimal",
			expected:      "minimal",
		},
		{
			name:          "Full variant",
			managedName:   "ocm-container",
			componentName: "full",
			expected:      "full",
		},
		{
			name:          "Not in ocm-container",
			managedName:   "",
			componentName: "full",
			expected:      "",
		},
		{
			name:          "Wrong managed name",
			managedName:   "some-other-container",
			componentName: "micro",
			expected:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("IO_OPENSHIFT_MANAGED_NAME", tt.managedName)
			t.Setenv("IO_OPENSHIFT_MANAGED_COMPONENT", tt.componentName)
			got := GetOcmContainerComponent()
			if got != tt.expected {
				t.Errorf("GetOcmContainerComponent() = %v, want %v", got, tt.expected)
			}
		})
	}
}
