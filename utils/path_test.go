package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "expands tilde path",
			input:    "~/Documents",
			expected: filepath.Join(home, "Documents"),
		},
		{
			name:     "expands nested tilde path",
			input:    "~/Library/Developer/Xcode",
			expected: filepath.Join(home, "Library/Developer/Xcode"),
		},
		{
			name:     "returns absolute path unchanged",
			input:    "/usr/local/bin",
			expected: "/usr/local/bin",
		},
		{
			name:     "returns relative path unchanged",
			input:    "relative/path",
			expected: "relative/path",
		},
		{
			name:     "handles tilde only",
			input:    "~/",
			expected: home,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandPath(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
