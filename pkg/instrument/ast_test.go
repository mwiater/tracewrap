package instrument_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mwiater/tracewrap/config"
	"github.com/mwiater/tracewrap/pkg/instrument"
)

func TestASTInstrumentation(t *testing.T) {
	// Create a temporary directory for our dummy Go file.
	tempDir, err := os.MkdirTemp("", "asttest")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Write a simple Go source file into the temporary directory.
	// This file defines one function with no instrumentation.
	dummySrc := `package main

func Hello() string {
	return "hello"
}
`
	dummyFile := filepath.Join(tempDir, "dummy.go")
	if err := os.WriteFile(dummyFile, []byte(dummySrc), 0644); err != nil {
		t.Fatalf("Failed to write dummy go file: %v", err)
	}

	// Set the dynamic tracer import so that the instrumentation process can add the proper import.
	if err := instrument.SetDynamicTracerImport(tempDir); err != nil {
		t.Fatalf("SetDynamicTracerImport failed: %v", err)
	}

	// Create a dummy configuration with no excludes.
	dummyConfig := config.Config{
		Instrumentation: config.InstrumentationConfig{
			Enable:  true,
			Include: []string{},
			Exclude: []string{},
		},
	}

	// Run InstrumentWorkspace on our temporary directory.
	// This will process all .go files (including dummy.go) in tempDir.
	if err := instrument.InstrumentWorkspace(tempDir, dummyConfig); err != nil {
		t.Fatalf("InstrumentWorkspace returned error: %v", err)
	}

	// Read back the dummy file and check that instrumentation was injected.
	data, err := os.ReadFile(dummyFile)
	if err != nil {
		t.Fatalf("Failed to read instrumented file: %v", err)
	}
	content := string(data)

	// Verify that a tracer call (e.g., RecordEntry) is present in the file.
	if !strings.Contains(content, "RecordEntry(") {
		t.Errorf("Instrumented file does not contain tracer call 'RecordEntry'; content: %s", content)
	}
}
