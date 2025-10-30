package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rdark/za/internal/config"
	"github.com/rdark/za/internal/notes"
)

func TestStandupSlack_WithBothDays(t *testing.T) {
	tempDir := t.TempDir()
	journalDir := filepath.Join(tempDir, "journal")

	if err := os.MkdirAll(journalDir, 0755); err != nil {
		t.Fatalf("failed to create journal dir: %v", err)
	}

	// Create yesterday's journal with completed work
	yesterday := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)
	yesterdayPath := filepath.Join(journalDir, yesterday.Format(notes.DateFormat)+".md")
	yesterdayContent := `---
title: Yesterday's Journal
---

## Goals of the Day

* [x] Complete feature X
* [ ] Review PR #123
* [x] Deploy to staging

# Work Completed

* Fixed bug in authentication
* Updated API documentation
`
	if err := os.WriteFile(yesterdayPath, []byte(yesterdayContent), 0644); err != nil {
		t.Fatalf("failed to create yesterday's journal: %v", err)
	}

	// Create today's journal with goals
	today := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)
	todayPath := filepath.Join(journalDir, today.Format(notes.DateFormat)+".md")
	todayContent := `---
title: Today's Journal
---

## Goals of the Day

* [ ] Review code changes
* [ ] Test new feature
* [x] Update documentation
`
	if err := os.WriteFile(todayPath, []byte(todayContent), 0644); err != nil {
		t.Fatalf("failed to create today's journal: %v", err)
	}

	// Configure
	cfg = &config.Config{
		Journal: config.JournalConfig{
			Dir:              journalDir,
			WorkDoneSections: []string{"Work Completed"},
		},
		SearchWindowDays: 30,
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command
	err := runStandupSlack(nil, []string{today.Format(notes.DateFormat)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout
	outputBytes, _ := io.ReadAll(r)
	output := string(outputBytes)

	// Verify output format
	if !strings.Contains(output, "previous:") {
		t.Error("expected output to contain 'previous:'")
	}
	if !strings.Contains(output, "next:") {
		t.Error("expected output to contain 'next:'")
	}

	// Verify yesterday's completed goals are included
	if !strings.Contains(output, "Complete feature X") {
		t.Error("expected output to contain completed goal from yesterday")
	}
	if !strings.Contains(output, "Deploy to staging") {
		t.Error("expected output to contain completed goal from yesterday")
	}

	// Verify yesterday's work items are included
	if !strings.Contains(output, "Fixed bug in authentication") {
		t.Error("expected output to contain work item from yesterday")
	}
	if !strings.Contains(output, "Updated API documentation") {
		t.Error("expected output to contain work item from yesterday")
	}

	// Verify uncompleted goals from yesterday are NOT included
	if strings.Contains(output, "Review PR #123") {
		t.Error("expected output to NOT contain uncompleted goal from yesterday")
	}

	// Verify today's goals are included (all goals, completed or not)
	if !strings.Contains(output, "Review code changes") {
		t.Error("expected output to contain today's goal")
	}
	if !strings.Contains(output, "Test new feature") {
		t.Error("expected output to contain today's goal")
	}
	if !strings.Contains(output, "Update documentation") {
		t.Error("expected output to contain today's goal")
	}

	// Verify format uses asterisks for bullets
	lines := strings.Split(output, "\n")
	var bulletCount int
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "* ") {
			bulletCount++
		}
	}
	if bulletCount < 5 {
		t.Errorf("expected at least 5 bullet points, got %d", bulletCount)
	}
}

func TestStandupSlack_NoYesterdayWork(t *testing.T) {
	tempDir := t.TempDir()
	journalDir := filepath.Join(tempDir, "journal")

	if err := os.MkdirAll(journalDir, 0755); err != nil {
		t.Fatalf("failed to create journal dir: %v", err)
	}

	// Only create today's journal
	today := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)
	todayPath := filepath.Join(journalDir, today.Format(notes.DateFormat)+".md")
	todayContent := `---
title: Today's Journal
---

## Goals of the Day

* [ ] Review code changes
`
	if err := os.WriteFile(todayPath, []byte(todayContent), 0644); err != nil {
		t.Fatalf("failed to create today's journal: %v", err)
	}

	cfg = &config.Config{
		Journal: config.JournalConfig{
			Dir:              journalDir,
			WorkDoneSections: []string{"Work Completed"},
		},
		SearchWindowDays: 30,
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runStandupSlack(nil, []string{today.Format(notes.DateFormat)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout
	outputBytes, _ := io.ReadAll(r)
	output := string(outputBytes)

	// Should still have both sections
	if !strings.Contains(output, "previous:") {
		t.Error("expected output to contain 'previous:'")
	}
	if !strings.Contains(output, "next:") {
		t.Error("expected output to contain 'next:'")
	}

	// Should indicate no work recorded
	if !strings.Contains(output, "No work recorded") {
		t.Error("expected output to indicate no work recorded")
	}

	// Should contain today's goals
	if !strings.Contains(output, "Review code changes") {
		t.Error("expected output to contain today's goal")
	}
}

func TestStandupSlack_NoTodayGoals(t *testing.T) {
	tempDir := t.TempDir()
	journalDir := filepath.Join(tempDir, "journal")

	if err := os.MkdirAll(journalDir, 0755); err != nil {
		t.Fatalf("failed to create journal dir: %v", err)
	}

	// Only create yesterday's journal
	yesterday := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)
	yesterdayPath := filepath.Join(journalDir, yesterday.Format(notes.DateFormat)+".md")
	yesterdayContent := `---
title: Yesterday's Journal
---

# Work Completed

* Fixed a bug
`
	if err := os.WriteFile(yesterdayPath, []byte(yesterdayContent), 0644); err != nil {
		t.Fatalf("failed to create yesterday's journal: %v", err)
	}

	today := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)

	cfg = &config.Config{
		Journal: config.JournalConfig{
			Dir:              journalDir,
			WorkDoneSections: []string{"Work Completed"},
		},
		SearchWindowDays: 30,
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runStandupSlack(nil, []string{today.Format(notes.DateFormat)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout
	outputBytes, _ := io.ReadAll(r)
	output := string(outputBytes)

	// Should have yesterday's work
	if !strings.Contains(output, "Fixed a bug") {
		t.Error("expected output to contain yesterday's work")
	}

	// Should indicate no goals set for today
	if !strings.Contains(output, "No goals set") {
		t.Error("expected output to indicate no goals set")
	}
}
