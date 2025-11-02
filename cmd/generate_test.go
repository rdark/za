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

	// Create a standup entry with structured sections
	standupDate := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)
	standupPath := filepath.Join(standupDir, standupDate.Format(notes.DateFormat)+".md")
	standupContent := `---
title: Standup
---

## Worked on yesterday

## Working on Today

## Notes

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

	// Should contain extracted work IN the "Worked on yesterday" section
	if !strings.Contains(contentStr, "Implemented feature X") {
		t.Error("expected standup to contain extracted work")
	}
	if !strings.Contains(contentStr, "Fixed bug Y") {
		t.Error("expected standup to contain extracted work")
	}

	// Verify work is in the correct section (between "Worked on yesterday" and "Working on Today")
	workedOnIdx := strings.Index(contentStr, "Worked on yesterday")
	workingOnIdx := strings.Index(contentStr, "Working on Today")
	featureIdx := strings.Index(contentStr, "Implemented feature X")

	if workedOnIdx == -1 || workingOnIdx == -1 || featureIdx == -1 {
		t.Errorf("missing sections or content: workedOn=%d, workingOn=%d, feature=%d",
			workedOnIdx, workingOnIdx, featureIdx)
	} else if featureIdx < workedOnIdx || featureIdx > workingOnIdx {
		t.Errorf("work content not in correct section: workedOn=%d, workingOn=%d, feature=%d",
			workedOnIdx, workingOnIdx, featureIdx)
	}

	// Should NOT contain content from other sections
	if strings.Contains(contentStr, "Some other content") {
		t.Error("expected standup to NOT contain content from other sections")
	}
}

func TestPopulateStandupWithWork_WithCompletedGoals(t *testing.T) {
	tempDir := t.TempDir()
	journalDir := filepath.Join(tempDir, "journal")
	standupDir := filepath.Join(tempDir, "standup")

	if err := os.MkdirAll(journalDir, 0755); err != nil {
		t.Fatalf("failed to create journal dir: %v", err)
	}
	if err := os.MkdirAll(standupDir, 0755); err != nil {
		t.Fatalf("failed to create standup dir: %v", err)
	}

	// Create a previous journal with work sections AND completed goals
	previousDate := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)
	journalPath := filepath.Join(journalDir, previousDate.Format(notes.DateFormat)+".md")
	journalContent := `---
title: Previous Journal
---

## Goals of the Day

* [x] Complete feature X implementation
* [ ] Review PR #123
* [x] Deploy to staging
* Plain bullet without checkbox

# Work Completed

* Fixed bug Y
* Updated documentation

# Other Section

Some other content
`
	if err := os.WriteFile(journalPath, []byte(journalContent), 0644); err != nil {
		t.Fatalf("failed to create journal: %v", err)
	}

	// Create a standup entry with structured sections
	standupDate := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)
	standupPath := filepath.Join(standupDir, standupDate.Format(notes.DateFormat)+".md")
	standupContent := `---
title: Standup
---

## Worked on yesterday

## Working on Today

## Notes

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

	// Should contain completed goals
	if !strings.Contains(contentStr, "Complete feature X implementation") {
		t.Error("expected standup to contain completed goal: Complete feature X implementation")
	}
	if !strings.Contains(contentStr, "Deploy to staging") {
		t.Error("expected standup to contain completed goal: Deploy to staging")
	}

	// Should NOT contain uncompleted goals
	if strings.Contains(contentStr, "Review PR #123") {
		t.Error("expected standup to NOT contain uncompleted goal: Review PR #123")
	}

	// Should NOT contain plain bullets without checkboxes
	if strings.Contains(contentStr, "Plain bullet without checkbox") {
		t.Error("expected standup to NOT contain plain bullet items")
	}

	// Should contain extracted work from Work Completed section
	if !strings.Contains(contentStr, "Fixed bug Y") {
		t.Error("expected standup to contain extracted work: Fixed bug Y")
	}
	if !strings.Contains(contentStr, "Updated documentation") {
		t.Error("expected standup to contain extracted work: Updated documentation")
	}

	// Verify work is in the correct section
	workedOnIdx := strings.Index(contentStr, "Worked on yesterday")
	workingOnIdx := strings.Index(contentStr, "Working on Today")
	featureIdx := strings.Index(contentStr, "Complete feature X implementation")

	if workedOnIdx == -1 || workingOnIdx == -1 || featureIdx == -1 {
		t.Errorf("missing sections or content: workedOn=%d, workingOn=%d, feature=%d",
			workedOnIdx, workingOnIdx, featureIdx)
	} else if featureIdx < workedOnIdx || featureIdx > workingOnIdx {
		t.Errorf("completed goals not in correct section: workedOn=%d, workingOn=%d, feature=%d",
			workedOnIdx, workingOnIdx, featureIdx)
	}

	// Verify no extra blank line between completed goals and work items
	// Goals end with "Deploy to staging\n", work starts with "* Fixed bug Y"
	deployIdx := strings.Index(contentStr, "Deploy to staging")
	bugIdx := strings.Index(contentStr, "Fixed bug Y")
	if deployIdx != -1 && bugIdx != -1 {
		betweenContent := contentStr[deployIdx+len("Deploy to staging") : bugIdx]
		// Should only have one newline, not multiple
		if strings.Count(betweenContent, "\n") > 1 {
			t.Errorf("expected no blank line between completed goals and work items, got: %q", betweenContent)
		}
	}

	// Should NOT contain content from other sections
	if strings.Contains(contentStr, "Some other content") {
		t.Error("expected standup to NOT contain content from other sections")
	}
}

func TestPopulateStandupWithWork_WithTodayGoals(t *testing.T) {
	tempDir := t.TempDir()
	journalDir := filepath.Join(tempDir, "journal")
	standupDir := filepath.Join(tempDir, "standup")

	if err := os.MkdirAll(journalDir, 0755); err != nil {
		t.Fatalf("failed to create journal dir: %v", err)
	}
	if err := os.MkdirAll(standupDir, 0755); err != nil {
		t.Fatalf("failed to create standup dir: %v", err)
	}

	// Create a previous journal with work
	previousDate := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)
	prevJournalPath := filepath.Join(journalDir, previousDate.Format(notes.DateFormat)+".md")
	prevJournalContent := `---
title: Previous Journal
---

# Work Completed

* Completed task A
`
	if err := os.WriteFile(prevJournalPath, []byte(prevJournalContent), 0644); err != nil {
		t.Fatalf("failed to create previous journal: %v", err)
	}

	// Create today's journal with goals
	standupDate := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)
	todayJournalPath := filepath.Join(journalDir, standupDate.Format(notes.DateFormat)+".md")
	todayJournalContent := `---
title: Today's Journal
---

## Goals of the Day

* [ ] Review code changes
* [x] Update documentation
* [ ] Test new feature
`
	if err := os.WriteFile(todayJournalPath, []byte(todayJournalContent), 0644); err != nil {
		t.Fatalf("failed to create today's journal: %v", err)
	}

	// Create a standup entry with structured sections
	standupPath := filepath.Join(standupDir, standupDate.Format(notes.DateFormat)+".md")
	standupContent := `---
title: Standup
---

## Worked on yesterday

## Working on Today

## Notes
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

	// Should contain yesterday's work in "Worked on yesterday" section
	if !strings.Contains(contentStr, "Completed task A") {
		t.Error("expected standup to contain yesterday's work")
	}

	// Should contain today's goals in "Working on Today" section
	if !strings.Contains(contentStr, "Review code changes") {
		t.Error("expected standup to contain today's goal: Review code changes")
	}
	if !strings.Contains(contentStr, "Update documentation") {
		t.Error("expected standup to contain today's goal: Update documentation")
	}
	if !strings.Contains(contentStr, "Test new feature") {
		t.Error("expected standup to contain today's goal: Test new feature")
	}

	// Verify today's goals are in the correct section
	workingOnIdx := strings.Index(contentStr, "Working on Today")
	notesIdx := strings.Index(contentStr, "Notes")
	reviewIdx := strings.Index(contentStr, "Review code changes")

	if workingOnIdx == -1 || notesIdx == -1 || reviewIdx == -1 {
		t.Errorf("missing sections or content: workingOn=%d, notes=%d, review=%d",
			workingOnIdx, notesIdx, reviewIdx)
	} else if reviewIdx < workingOnIdx || reviewIdx > notesIdx {
		t.Errorf("today's goals not in correct section: workingOn=%d, notes=%d, review=%d",
			workingOnIdx, notesIdx, reviewIdx)
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
	if err != nil {
		t.Fatalf("expected no error when no previous journal exists, got: %v", err)
	}

	// Verify standup file was not modified (since there's no work to populate)
	content, err := os.ReadFile(standupPath)
	if err != nil {
		t.Fatalf("failed to read standup file: %v", err)
	}
	if string(content) != standupContent {
		t.Errorf("standup file was unexpectedly modified: got %q, want %q", string(content), standupContent)
	}
}
