package instrument_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mwiater/tracewrap/pkg/instrument"
)

func TestPrepareWorkspace(t *testing.T) {
	// Create a temporary source directory with a sample structure.
	srcDir, err := os.MkdirTemp("", "prepareworkspace-src")
	if err != nil {
		t.Fatalf("Failed to create source temp directory: %v", err)
	}
	// Clean up the source directory after the test.
	defer os.RemoveAll(srcDir)

	// Create a file in the source directory.
	filePath := filepath.Join(srcDir, "file.txt")
	fileContent := []byte("sample content")
	if err := os.WriteFile(filePath, fileContent, 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Create a subdirectory with a file.
	subDir := filepath.Join(srcDir, "sub")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create sub directory: %v", err)
	}
	subFilePath := filepath.Join(subDir, "subfile.txt")
	subFileContent := []byte("sub file content")
	if err := os.WriteFile(subFilePath, subFileContent, 0644); err != nil {
		t.Fatalf("Failed to create sub file: %v", err)
	}

	// Call PrepareWorkspace to copy the source directory into a new temporary workspace.
	workspaceDir, err := instrument.PrepareWorkspace(srcDir)
	if err != nil {
		t.Fatalf("PrepareWorkspace returned error: %v", err)
	}
	// Clean up the workspace directory after the test.
	defer os.RemoveAll(workspaceDir)

	// Verify that the file from the root of the source was copied.
	copiedFilePath := filepath.Join(workspaceDir, "file.txt")
	data, err := os.ReadFile(copiedFilePath)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}
	if string(data) != string(fileContent) {
		t.Errorf("Copied file content mismatch: expected %q, got %q", string(fileContent), string(data))
	}

	// Verify that the file in the subdirectory was copied.
	copiedSubFilePath := filepath.Join(workspaceDir, "sub", "subfile.txt")
	subData, err := os.ReadFile(copiedSubFilePath)
	if err != nil {
		t.Fatalf("Failed to read copied sub file: %v", err)
	}
	if string(subData) != string(subFileContent) {
		t.Errorf("Copied sub file content mismatch: expected %q, got %q", string(subFileContent), string(subData))
	}
}
