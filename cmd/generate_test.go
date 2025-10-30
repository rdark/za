package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rdark/za/internal/config"
	"github.com/rdark/za/internal/notes"
)

func TestGenerateJournal_MissingConfig(t *testing.T) {
	// Create temp config with empty create command
	tempDir := t.TempDir()
	cfg = &config.Config{
		Journal: config.JournalConfig{
			Dir:                tempDir,
			WorkDoneSections:   []string{"work completed"},
			LinkPreviousTitles: []string{"Yesterday"},
			LinkNextTitles:     []string{"Tomorrow"},
			Create:             config.CreateCommand{Cmd: ""}, // Empty command
		},
		SearchWindowDays: 30,
	}

	err := runGenerateJournal(nil, []string{})
	if err == nil {
		t.Fatal("expected error for missing create command, got nil")
	}

	if !strings.Contains(err.Error(), "not configured") {
		t.Errorf("expected 'not configured' error, got: %v", err)
	}
}

func TestGenerateJournal_InvalidDate(t *testing.T) {
	tempDir := t.TempDir()
	cfg = &config.Config{
		Journal: config.JournalConfig{
			Dir:    tempDir,
			Create: config.CreateCommand{Cmd: "echo test"},
		},
		SearchWindowDays: 30,
	}

	err := runGenerateJournal(nil, []string{"invalid-date"})
	if err == nil {
		t.Fatal("expected error for invalid date, got nil")
	}

	if !strings.Contains(err.Error(), "invalid date format") {
		t.Errorf("expected 'invalid date format' error, got: %v", err)
	}
}

func TestGenerateJournal_FileAlreadyExists(t *testing.T) {
	tempDir := t.TempDir()

	// Create a journal entry that already exists
	dateStr := "2025-01-15"
	existingFile := filepath.Join(tempDir, dateStr+".md")
	if err := os.WriteFile(existingFile, []byte("existing content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cfg = &config.Config{
		Journal: config.JournalConfig{
			Dir:                tempDir,
			WorkDoneSections:   []string{"work completed"},
			LinkPreviousTitles: []string{"Yesterday"},
			LinkNextTitles:     []string{"Tomorrow"},
			Create:             config.CreateCommand{Cmd: "echo test > " + existingFile},
		},
		SearchWindowDays: 30,
	}

	err := runGenerateJournal(nil, []string{dateStr})
	if err == nil {
		t.Fatal("expected error for existing file, got nil")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestGenerateJournal_Success(t *testing.T) {
	tempDir := t.TempDir()
	dateStr := "2025-01-20"
	targetFile := filepath.Join(tempDir, dateStr+".md")

	// Create command that creates a file
	createCmd := "echo '# Test Journal' > " + targetFile

	cfg = &config.Config{
		Journal: config.JournalConfig{
			Dir:                tempDir,
			WorkDoneSections:   []string{"work completed"},
			LinkPreviousTitles: []string{"Yesterday"},
			LinkNextTitles:     []string{"Tomorrow"},
			Create:             config.CreateCommand{Cmd: createCmd},
		},
		SearchWindowDays: 30,
	}

	// Suppress output for test
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	err := runGenerateJournal(nil, []string{dateStr})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		t.Errorf("expected file to be created at %s", targetFile)
	}

	// Verify content
	content, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	if !strings.Contains(string(content), "Test Journal") {
		t.Errorf("expected file to contain 'Test Journal', got: %s", string(content))
	}
}

func TestGenerateStandup_MissingConfig(t *testing.T) {
	tempDir := t.TempDir()
	cfg = &config.Config{
		Standup: config.StandupConfig{
			Dir:                tempDir,
			WorkDoneSection:    "Worked on yesterday",
			LinkPreviousTitles: []string{"Yesterday"},
			LinkNextTitles:     []string{"Tomorrow"},
			Create:             config.CreateCommand{Cmd: ""}, // Empty command
		},
		SearchWindowDays: 30,
	}

	err := runGenerateStandup(nil, []string{})
	if err == nil {
		t.Fatal("expected error for missing create command, got nil")
	}

	if !strings.Contains(err.Error(), "not configured") {
		t.Errorf("expected 'not configured' error, got: %v", err)
	}
}

func TestGenerateStandup_InvalidDate(t *testing.T) {
	tempDir := t.TempDir()
	cfg = &config.Config{
		Standup: config.StandupConfig{
			Dir:    tempDir,
			Create: config.CreateCommand{Cmd: "echo test"},
		},
		SearchWindowDays: 30,
	}

	err := runGenerateStandup(nil, []string{"not-a-date"})
	if err == nil {
		t.Fatal("expected error for invalid date, got nil")
	}

	if !strings.Contains(err.Error(), "invalid date format") {
		t.Errorf("expected 'invalid date format' error, got: %v", err)
	}
}

func TestGenerateStandup_Success(t *testing.T) {
	tempDir := t.TempDir()
	standupDir := filepath.Join(tempDir, "standup")
	if err := os.MkdirAll(standupDir, 0755); err != nil {
		t.Fatalf("failed to create standup dir: %v", err)
	}

	dateStr := "2025-01-21"
	targetFile := filepath.Join(standupDir, dateStr+".md")

	// Create command that creates a file
	createCmd := "echo '# Standup' > " + targetFile

	cfg = &config.Config{
		Standup: config.StandupConfig{
			Dir:                standupDir,
			WorkDoneSection:    "Worked on yesterday",
			LinkPreviousTitles: []string{"Yesterday"},
			LinkNextTitles:     []string{"Tomorrow"},
			Create:             config.CreateCommand{Cmd: createCmd},
		},
		SearchWindowDays: 30,
	}

	// Suppress output for test
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	err := runGenerateStandup(nil, []string{dateStr})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		t.Errorf("expected file to be created at %s", targetFile)
	}

	// Verify content
	content, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	if !strings.Contains(string(content), "Standup") {
		t.Errorf("expected file to contain 'Standup', got: %s", string(content))
	}
}

func TestPopulateStandupWithWork(t *testing.T) {
	tempDir := t.TempDir()
	journalDir := filepath.Join(tempDir, "journal")
	standupDir := filepath.Join(tempDir, "standup")

	if err := os.MkdirAll(journalDir, 0755); err != nil {
		t.Fatalf("failed to create journal dir: %v", err)
	}
	if err := os.MkdirAll(standupDir, 0755); err != nil {
		t.Fatalf("failed to create standup dir: %v", err)
	}

	// Create a previous journal entry with work
	previousDate := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)
	journalPath := filepath.Join(journalDir, previousDate.Format(notes.DateFormat)+".md")
	journalContent := `---
title: Previous Journal
---

# Work Completed

* Implemented feature X
* Fixed bug Y

# Other Section

Some other content
`
	if err := os.WriteFile(journalPath, []byte(journalContent), 0644); err != nil {
		t.Fatalf("failed to create journal: %v", err)
	}

	// Create an empty standup entry
	standupDate := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)
	standupPath := filepath.Join(standupDir, standupDate.Format(notes.DateFormat)+".md")
	standupContent := `---
title: Standup
---

# Summary

This is my standup
`
	if err := os.WriteFile(standupPath, []byte(standupContent), 0644); err != nil {
		t.Fatalf("failed to create standup: %v", err)
	}

	// Configure
	cfg = &config.Config{
		Journal: config.JournalConfig{
			Dir:              journalDir,
			WorkDoneSections: []string{"Work Completed"},
		},
		Standup: config.StandupConfig{
			Dir:             standupDir,
			WorkDoneSection: "Worked on yesterday",
		},
		SearchWindowDays: 30,
	}

	// Suppress output for test
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	// Populate standup with work
	err := populateStandupWithWork(standupDate, standupPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify standup was updated
	updatedContent, err := os.ReadFile(standupPath)
	if err != nil {
		t.Fatalf("failed to read updated standup: %v", err)
	}

	contentStr := string(updatedContent)

	// Should contain original content
	if !strings.Contains(contentStr, "This is my standup") {
		t.Error("expected standup to contain original content")
	}

	// Should contain work section header
	if !strings.Contains(contentStr, "Worked on yesterday") {
		t.Error("expected standup to contain work section header")
	}

	// Should contain extracted work
	if !strings.Contains(contentStr, "Implemented feature X") {
		t.Error("expected standup to contain extracted work")
	}
	if !strings.Contains(contentStr, "Fixed bug Y") {
		t.Error("expected standup to contain extracted work")
	}

	// Should NOT contain content from other sections
	if strings.Contains(contentStr, "Some other content") {
		t.Error("expected standup to NOT contain content from other sections")
	}
}

func TestPopulateStandupWithWork_NoPreviousJournal(t *testing.T) {
	tempDir := t.TempDir()
	journalDir := filepath.Join(tempDir, "journal")
	standupDir := filepath.Join(tempDir, "standup")

	if err := os.MkdirAll(journalDir, 0755); err != nil {
		t.Fatalf("failed to create journal dir: %v", err)
	}
	if err := os.MkdirAll(standupDir, 0755); err != nil {
		t.Fatalf("failed to create standup dir: %v", err)
	}

	// Create standup but NO previous journal
	standupDate := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)
	standupPath := filepath.Join(standupDir, standupDate.Format(notes.DateFormat)+".md")
	standupContent := "# Standup\n"
	if err := os.WriteFile(standupPath, []byte(standupContent), 0644); err != nil {
		t.Fatalf("failed to create standup: %v", err)
	}

	cfg = &config.Config{
		Journal: config.JournalConfig{
			Dir:              journalDir,
			WorkDoneSections: []string{"Work Completed"},
		},
		Standup: config.StandupConfig{
			Dir:             standupDir,
			WorkDoneSection: "Worked on yesterday",
		},
		SearchWindowDays: 30,
	}

	// Suppress output for test
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	err := populateStandupWithWork(standupDate, standupPath)
	if err == nil {
		t.Fatal("expected error when no previous journal exists")
	}

	if !strings.Contains(err.Error(), "could not find previous journal") {
		t.Errorf("expected 'could not find previous journal' error, got: %v", err)
	}
}
