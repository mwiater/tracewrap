// tests/e2e_test.go
package e2e_test

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// copyDir is a helper function that copies the contents of the src directory into dst.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(targetPath, data, info.Mode())
	})
}

func TestSimpleExampleEndToEnd(t *testing.T) {
	// Create a temporary directory for the test.
	tempDir, err := os.MkdirTemp("", "tracewrap-e2e-simple")
	if err != nil {
		t.Fatal("Failed to create temp directory:", err)
	}
	defer os.RemoveAll(tempDir)

	// Determine the project root.
	// Since this test file is located in tests/, we assume the project root is one level up.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("Failed to get working directory:", err)
	}
	rootDir := filepath.Join(wd, "..")

	// Copy the simple example project from the project root's examples/simple folder into the temp directory.
	srcDir := filepath.Join(rootDir, "examples", "simple")
	dstDir := filepath.Join(tempDir, "simple")
	if err := copyDir(srcDir, dstDir); err != nil {
		t.Fatal("Failed to copy simple example project:", err)
	}

	// Update the go.mod in the copied project to add a replace directive,
	// so that github.com/mwiater/tracewrap resolves to our local project.
	modCmd := exec.Command("go", "mod", "edit", "-replace", "github.com/mwiater/tracewrap="+rootDir)
	modCmd.Dir = dstDir
	if output, err := modCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to set replace directive: %v, output: %s", err, string(output))
	}

	// Define the path to your tracewrap binary. It should be at <project-root>/bin/tracewrap.
	tracewrapBinary := filepath.Join(rootDir, "bin", "tracewrap")
	if _, err := os.Stat(tracewrapBinary); os.IsNotExist(err) {
		t.Fatal("tracewrap binary not found at", tracewrapBinary)
	}

	// Run the tracewrap instrumentation command on the copied project.
	cmd := exec.Command(tracewrapBinary,
		"buildTracedApplication",
		"--project", dstDir,
		"--config", filepath.Join(dstDir, "tracewrap.yaml"),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatal("Failed to run tracewrap instrumentation command:", err)
	}

	// Verify that the expected output files exist in the instrumented project.
	logFilePath := filepath.Join(dstDir, "tracewrap", "tracewrap.log")
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		t.Fatal("Expected log file does not exist:", logFilePath)
	}

	callGraphPath := filepath.Join(dstDir, "tracewrap", "callgraph.dot")
	if _, err := os.Stat(callGraphPath); os.IsNotExist(err) {
		t.Fatal("Expected call graph file does not exist:", callGraphPath)
	}

	// Optionally, read the log file to check for expected tracer output.
	logContent, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatal("Failed to read log file:", err)
	}
	if !strings.Contains(string(logContent), "Entering") {
		t.Errorf("Log file does not appear to contain tracer output. Got: %s", string(logContent))
	}
}
