package main_test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestMainNoProject(t *testing.T) {
	// Construct the command that runs the tracewrap binary.
	// Since your file tree places the binary in the "bin" directory at the project root,
	// and this test file is in "tests/main/", we reference the binary as "../../bin/tracewrap".
	cmd := exec.Command("../../bin/tracewrap", "buildTracedApplication")

	// Run the command and capture its combined standard output and standard error.
	output, err := cmd.CombinedOutput()
	outStr := string(output)

	// We expect the command to exit with an error since the --project flag is required.
	if err == nil {
		t.Fatalf("Expected nonzero exit status when no project flag is provided, but command succeeded. Output:\n%s", outStr)
	}

	// Check that the expected error message is present.
	expectedMessage := "Project directory must be specified using --project"
	if !strings.Contains(outStr, expectedMessage) {
		t.Errorf("Expected output to contain %q, but got:\n%s", expectedMessage, outStr)
	}
}
