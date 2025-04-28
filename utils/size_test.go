package utils

import (
	"os"
	"testing"
)

func TestCalculateDirSize(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a dummy file
	filePath := tmpDir + "/testfile"
	content := []byte("Hello World")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	expectedSize := int64(len(content))
	size := CalculateDirSize(tmpDir)
	if size != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, size)
	}
}
