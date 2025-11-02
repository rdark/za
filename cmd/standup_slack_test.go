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
	standupDir := filepath.Join(tempDir, "standup")

	if err := os.MkdirAll(standupDir, 0755); err != nil {
		t.Fatalf("failed to create standup dir: %v", err)
	}

	// Create today's standup with work from yesterday and goals for today
	today := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)
	standupPath := filepath.Join(standupDir, today.Format(notes.DateFormat)+".md")
	standupContent := `---
title: Standup 2025-01-21
---

# Standup 2025-01-21

## Worked on Yesterday

* [Yesterday](../journal/2025-01-20)

* Complete feature X
* Deploy to staging
* Fixed bug in authentication
* Updated API documentation

## Working on Today

* [Today](../journal/2025-01-21)

* Review code changes
* Test new feature
* Update documentation
`
	if err := os.WriteFile(standupPath, []byte(standupContent), 0644); err != nil {
		t.Fatalf("failed to create standup: %v", err)
	}

	// Configure
	cfg = &config.Config{
		Standup: config.StandupConfig{
			Dir:             standupDir,
			WorkDoneSection: "Worked on Yesterday",
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
	standupDir := filepath.Join(tempDir, "standup")

	if err := os.MkdirAll(standupDir, 0755); err != nil {
		t.Fatalf("failed to create standup dir: %v", err)
	}

	// Create today's standup with empty "Worked on Yesterday" section (only nav links) and today's goals
	today := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)
	standupPath := filepath.Join(standupDir, today.Format(notes.DateFormat)+".md")
	standupContent := `---
title: Standup 2025-01-21
---

# Standup 2025-01-21

## Worked on Yesterday

* [Yesterday](../journal/2025-01-20)

## Working on Today

* [Today](../journal/2025-01-21)

* Review code changes
`
	if err := os.WriteFile(standupPath, []byte(standupContent), 0644); err != nil {
		t.Fatalf("failed to create standup: %v", err)
	}

	cfg = &config.Config{
		Standup: config.StandupConfig{
			Dir:             standupDir,
			WorkDoneSection: "Worked on Yesterday",
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
	standupDir := filepath.Join(tempDir, "standup")

	if err := os.MkdirAll(standupDir, 0755); err != nil {
		t.Fatalf("failed to create standup dir: %v", err)
	}

	// Create today's standup with yesterday's work but empty "Working on Today" section (only nav links)
	today := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)
	standupPath := filepath.Join(standupDir, today.Format(notes.DateFormat)+".md")
	standupContent := `---
title: Standup 2025-01-21
---

# Standup 2025-01-21

## Worked on Yesterday

* [Yesterday](../journal/2025-01-20)

* Fixed a bug

## Working on Today

* [Today](../journal/2025-01-21)
`
	if err := os.WriteFile(standupPath, []byte(standupContent), 0644); err != nil {
		t.Fatalf("failed to create standup: %v", err)
	}

	cfg = &config.Config{
		Standup: config.StandupConfig{
			Dir:             standupDir,
			WorkDoneSection: "Worked on Yesterday",
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
