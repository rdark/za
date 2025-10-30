package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rdark/za/internal/config"
)

func TestGenerateConfig_Success(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test-config.yaml")

	// Set config output
	configOutput = outputFile
	configForce = false
	configMinimal = false

	// Suppress stdout
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	err := runGenerateConfig(nil, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("config file was not created: %s", outputFile)
	}

	// Verify content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	contentStr := string(content)
	expectedStrings := []string{
		"journal:",
		"standup:",
		"dir:",
		"work_done_sections:",
		"search_window_days:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("config file missing expected string: %s", expected)
		}
	}
}

func TestGenerateConfig_FileExists(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "existing.yaml")

	// Create existing file
	if err := os.WriteFile(outputFile, []byte("existing content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	configOutput = outputFile
	configForce = false

	err := runGenerateConfig(nil, []string{})
	if err == nil {
		t.Fatal("expected error when file exists, got nil")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestGenerateConfig_ForceOverwrite(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "existing.yaml")

	// Create existing file
	if err := os.WriteFile(outputFile, []byte("old content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	configOutput = outputFile
	configForce = true
	configMinimal = false

	// Suppress stdout
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	err := runGenerateConfig(nil, []string{})
	if err != nil {
		t.Fatalf("unexpected error with --force: %v", err)
	}

	// Verify file was overwritten
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	if strings.Contains(string(content), "old content") {
		t.Error("file was not overwritten")
	}

	if !strings.Contains(string(content), "journal:") {
		t.Error("new content was not written")
	}
}

func TestGenerateConfig_Minimal(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "minimal.yaml")

	configOutput = outputFile
	configForce = false
	configMinimal = true

	// Suppress stdout
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	err := runGenerateConfig(nil, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	contentStr := string(content)

	// Should have required fields
	if !strings.Contains(contentStr, "journal:") {
		t.Error("minimal config missing 'journal:'")
	}

	// Should NOT have comments
	if strings.Contains(contentStr, "# Za (Zettelkasten") {
		t.Error("minimal config should not contain comments")
	}

	// Verify it's more compact
	if len(contentStr) > 500 {
		t.Errorf("minimal config too large: %d bytes (expected < 500)", len(contentStr))
	}
}

func TestGenerateConfig_CustomOutput(t *testing.T) {
	tempDir := t.TempDir()
	customFile := filepath.Join(tempDir, "custom", "my-config.yaml")

	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(customFile), 0755); err != nil {
		t.Fatalf("failed to create parent dir: %v", err)
	}

	configOutput = customFile
	configForce = false
	configMinimal = true

	// Suppress stdout
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	err := runGenerateConfig(nil, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created at custom location
	if _, err := os.Stat(customFile); os.IsNotExist(err) {
		t.Errorf("config file was not created at custom path: %s", customFile)
	}
}

func TestGenerateConfig_ValidYAML(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "valid.yaml")

	configOutput = outputFile
	configForce = false
	configMinimal = false

	// Suppress stdout
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	err := runGenerateConfig(nil, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Try to load the generated config
	_, err = config.Load(outputFile)
	if err != nil {
		t.Errorf("generated config is not valid YAML: %v", err)
	}
}

func TestGenerateMinimalConfig(t *testing.T) {
	content := generateMinimalConfig()

	// Check it's actually minimal
	if strings.Contains(content, "#") {
		t.Error("minimal config should not contain comments")
	}

	// Check required fields
	requiredFields := []string{
		"journal:",
		"standup:",
		"dir:",
		"work_done_sections:",
		"search_window_days:",
	}

	for _, field := range requiredFields {
		if !strings.Contains(content, field) {
			t.Errorf("minimal config missing required field: %s", field)
		}
	}
}

func TestGenerateFullConfig(t *testing.T) {
	content := generateFullConfig()

	// Check it has comments
	if !strings.Contains(content, "# Za (Zettelkasten") {
		t.Error("full config should contain header comment")
	}

	// Check for example commands
	if !strings.Contains(content, "zk new") {
		t.Error("full config should contain example commands")
	}

	// Check it's reasonably sized (has explanations)
	if len(content) < 1000 {
		t.Errorf("full config seems too small: %d bytes", len(content))
	}

	// Check required sections
	requiredSections := []string{
		"# Journal Configuration",
		"# Standup Configuration",
		"# General Settings",
	}

	for _, section := range requiredSections {
		if !strings.Contains(content, section) {
			t.Errorf("full config missing section: %s", section)
		}
	}
}
