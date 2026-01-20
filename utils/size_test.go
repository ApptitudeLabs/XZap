package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCalculateDirSize(t *testing.T) {
	t.Run("single file", func(t *testing.T) {
		tmpDir := t.TempDir()
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
	})

	t.Run("multiple files", func(t *testing.T) {
		tmpDir := t.TempDir()
		content1 := []byte("File one content")
		content2 := []byte("File two content here")

		if err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), content1, 0644); err != nil {
			t.Fatalf("Failed to write file1: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "file2.txt"), content2, 0644); err != nil {
			t.Fatalf("Failed to write file2: %v", err)
		}

		expectedSize := int64(len(content1) + len(content2))
		size := CalculateDirSize(tmpDir)
		if size != expectedSize {
			t.Errorf("Expected size %d, got %d", expectedSize, size)
		}
	})

	t.Run("nested directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		nestedDir := filepath.Join(tmpDir, "subdir", "nested")
		if err := os.MkdirAll(nestedDir, 0755); err != nil {
			t.Fatalf("Failed to create nested dir: %v", err)
		}

		content1 := []byte("Root file")
		content2 := []byte("Nested file content")

		if err := os.WriteFile(filepath.Join(tmpDir, "root.txt"), content1, 0644); err != nil {
			t.Fatalf("Failed to write root file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(nestedDir, "nested.txt"), content2, 0644); err != nil {
			t.Fatalf("Failed to write nested file: %v", err)
		}

		expectedSize := int64(len(content1) + len(content2))
		size := CalculateDirSize(tmpDir)
		if size != expectedSize {
			t.Errorf("Expected size %d, got %d", expectedSize, size)
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		size := CalculateDirSize(tmpDir)
		if size != 0 {
			t.Errorf("Expected size 0 for empty dir, got %d", size)
		}
	})

	t.Run("non-existent directory", func(t *testing.T) {
		size := CalculateDirSize("/non/existent/path")
		if size != 0 {
			t.Errorf("Expected size 0 for non-existent dir, got %d", size)
		}
	})
}

func TestSplitSimulators(t *testing.T) {
	sims := []SimulatorInfo{
		{Name: "iPhone 15 Pro", Size: 5 << 30, UDID: "uuid-1"}, // 5 GB - critical
		{Name: "iPhone 14", Size: 2 << 30, UDID: "uuid-2"},     // 2 GB - normal
		{Name: "iPad Pro", Size: 4 << 30, UDID: "uuid-3"},      // 4 GB - critical
		{Name: "iPhone SE", Size: 500 << 20, UDID: "uuid-4"},   // 500 MB - normal
		{Name: "iPhone 13", Size: 3 << 30, UDID: "uuid-5"},     // 3 GB - exactly at threshold
	}

	t.Run("default threshold (3GB)", func(t *testing.T) {
		critical, normal := splitSimulators(sims, 0)

		if len(critical) != 3 {
			t.Errorf("Expected 3 critical sims, got %d", len(critical))
		}
		if len(normal) != 2 {
			t.Errorf("Expected 2 normal sims, got %d", len(normal))
		}

		// Verify critical are sorted by size descending
		if critical[0].Name != "iPhone 15 Pro" {
			t.Errorf("Expected iPhone 15 Pro first, got %s", critical[0].Name)
		}
		if critical[1].Name != "iPad Pro" {
			t.Errorf("Expected iPad Pro second, got %s", critical[1].Name)
		}
	})

	t.Run("custom threshold (4GB)", func(t *testing.T) {
		critical, normal := splitSimulators(sims, 4)

		if len(critical) != 2 {
			t.Errorf("Expected 2 critical sims with 4GB threshold, got %d", len(critical))
		}
		// With custom threshold, normal is not populated
		if len(normal) != 0 {
			t.Errorf("Expected 0 normal sims with custom threshold, got %d", len(normal))
		}
	})

	t.Run("empty input", func(t *testing.T) {
		critical, normal := splitSimulators([]SimulatorInfo{}, 0)

		if len(critical) != 0 {
			t.Errorf("Expected 0 critical sims, got %d", len(critical))
		}
		if len(normal) != 0 {
			t.Errorf("Expected 0 normal sims, got %d", len(normal))
		}
	})
}

func TestRuntimeFolderFromIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		expected   string
	}{
		{
			name:       "iOS runtime identifier",
			identifier: "com.apple.CoreSimulator.SimRuntime.iOS-17-5",
			expected:   "iOS-17-5.simruntime",
		},
		{
			name:       "watchOS runtime identifier",
			identifier: "com.apple.CoreSimulator.SimRuntime.watchOS-10-4",
			expected:   "watchOS-10-4.simruntime",
		},
		{
			name:       "tvOS runtime identifier",
			identifier: "com.apple.CoreSimulator.SimRuntime.tvOS-17-4",
			expected:   "tvOS-17-4.simruntime",
		},
		{
			name:       "simple identifier",
			identifier: "iOS-17-0",
			expected:   "iOS-17-0.simruntime",
		},
		{
			name:       "single part identifier",
			identifier: "runtime",
			expected:   "runtime.simruntime",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runtimeFolderFromIdentifier(tt.identifier)
			if result != tt.expected {
				t.Errorf("runtimeFolderFromIdentifier(%q) = %q, want %q", tt.identifier, result, tt.expected)
			}
		})
	}
}
