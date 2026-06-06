package docker

import (
	"testing"
)

func TestParseContainersList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty output",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single container",
			input:    "my-container",
			expected: []string{"my-container"},
		},
		{
			name:     "multiple containers",
			input:    "container1\ncontainer2\ncontainer3",
			expected: []string{"container1", "container2", "container3"},
		},
		{
			name:     "containers with whitespace",
			input:    "  container1  \n  container2  \n",
			expected: []string{"container1", "container2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseContainersList(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d containers, got %d", len(tt.expected), len(result))
				return
			}

			for i, container := range result {
				if container != tt.expected[i] {
					t.Errorf("expected container %s, got %s", tt.expected[i], container)
				}
			}
		})
	}
}

func TestParseContainersOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[int]string
	}{
		{
			name:     "empty output",
			input:    "",
			expected: map[int]string{},
		},
		{
			name:     "malformed output",
			input:    "incomplete",
			expected: map[int]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseContainersOutput(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d entries, got %d", len(tt.expected), len(result))
			}
		})
	}
}

func TestGetContainerForPort(t *testing.T) {
	// This is an integration test that requires Docker to be running
	// Skip if we can't connect to Docker
	result := GetContainerForPort(99999) // Use a port that's unlikely to exist
	if result != "" {
		t.Errorf("expected empty string for non-existent port, got %s", result)
	}
}
