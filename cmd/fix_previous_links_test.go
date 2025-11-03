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

func TestFixPreviousLinks_Journal(t *testing.T) {
	tempDir := t.TempDir()
	journalDir := filepath.Join(tempDir, "journal")

	if err := os.MkdirAll(journalDir, 0755); err != nil {
		t.Fatalf("failed to create journal dir: %v", err)
	}

	// Create previous day's journal with a "Tomorrow" link that has wrong date
	previousDate := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)
	prevJournalPath := filepath.Join(journalDir, previousDate.Format(notes.DateFormat)+".md")
	prevJournalContent := `---
title: Previous Journal
---

# Daily Log 2025-01-20

* [Yesterday](../journal/2025-01-19.md)
* [Tomorrow](../journal/2025-01-20.md)

## Goals of the Day

* [ ] Test goal
`
	if err := os.WriteFile(prevJournalPath, []byte(prevJournalContent), 0644); err != nil {
		t.Fatalf("failed to create previous journal: %v", err)
	}

	// Configure
	cfg = &config.Config{
		Journal: config.JournalConfig{
			Dir:              journalDir,
			LinkNextTitles:   []string{"Tomorrow"},
			WorkDoneSections: []string{"Work Completed"},
		},
		SearchWindowDays: 30,
	}

	// Call fixPreviousLinks for the current date (2025-01-21)
	currentDate := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)
	err := fixPreviousLinks(currentDate, notes.NoteTypeJournal, journalDir)
	if err != nil {
		t.Fatalf("fixPreviousLinks failed: %v", err)
	}

	// Read the updated previous journal
	updatedContent, err := os.ReadFile(prevJournalPath)
	if err != nil {
		t.Fatalf("failed to read updated journal: %v", err)
	}

	contentStr := string(updatedContent)

	// Verify the "Tomorrow" link was updated to point to 2025-01-21
	if !strings.Contains(contentStr, "[Tomorrow](../journal/2025-01-21.md)") {
		t.Errorf("expected Tomorrow link to be updated to 2025-01-21.md, got:\n%s", contentStr)
	}

	// Verify "Yesterday" link was not modified
	if !strings.Contains(contentStr, "[Yesterday](../journal/2025-01-19.md)") {
		t.Errorf("expected Yesterday link to remain unchanged, got:\n%s", contentStr)
	}
}

func TestFixPreviousLinks_Standup(t *testing.T) {
	tempDir := t.TempDir()
	standupDir := filepath.Join(tempDir, "standup")

	if err := os.MkdirAll(standupDir, 0755); err != nil {
		t.Fatalf("failed to create standup dir: %v", err)
	}

	// Create previous day's standup with "Next" link that has wrong date
	previousDate := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)
	prevStandupPath := filepath.Join(standupDir, previousDate.Format(notes.DateFormat)+".md")
	prevStandupContent := `---
title: Standup 2025-01-20
---

# Standup 2025-01-20

* [Previous](../standup/2025-01-19.md)
* [Next](../standup/2025-01-20.md)

## Worked on Yesterday

* Some work

## Working on Today

* Some goals
`
	if err := os.WriteFile(prevStandupPath, []byte(prevStandupContent), 0644); err != nil {
		t.Fatalf("failed to create previous standup: %v", err)
	}

	// Configure
	cfg = &config.Config{
		Standup: config.StandupConfig{
			Dir:            standupDir,
			LinkNextTitles: []string{"Next"},
		},
		SearchWindowDays: 30,
	}

	// Call fixPreviousLinks for the current date (2025-01-21)
	currentDate := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)
	err := fixPreviousLinks(currentDate, notes.NoteTypeStandup, standupDir)
	if err != nil {
		t.Fatalf("fixPreviousLinks failed: %v", err)
	}

	// Read the updated previous standup
	updatedContent, err := os.ReadFile(prevStandupPath)
	if err != nil {
		t.Fatalf("failed to read updated standup: %v", err)
	}

	contentStr := string(updatedContent)

	// Verify the "Next" link was updated to point to 2025-01-21
	if !strings.Contains(contentStr, "[Next](../standup/2025-01-21.md)") {
		t.Errorf("expected Next link to be updated to 2025-01-21.md, got:\n%s", contentStr)
	}

	// Verify "Previous" link was not modified
	if !strings.Contains(contentStr, "[Previous](../standup/2025-01-19.md)") {
		t.Errorf("expected Previous link to remain unchanged, got:\n%s", contentStr)
	}
}

func TestFixPreviousLinks_NoPreviousNote(t *testing.T) {
	tempDir := t.TempDir()
	journalDir := filepath.Join(tempDir, "journal")

	if err := os.MkdirAll(journalDir, 0755); err != nil {
		t.Fatalf("failed to create journal dir: %v", err)
	}

	// Configure
	cfg = &config.Config{
		Journal: config.JournalConfig{
			Dir: journalDir,
		},
		SearchWindowDays: 30,
	}

	// Call fixPreviousLinks when no previous note exists
	currentDate := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)
	err := fixPreviousLinks(currentDate, notes.NoteTypeJournal, journalDir)

	// Should not return error when no previous note exists
	if err != nil {
		t.Errorf("expected no error when no previous note exists, got: %v", err)
	}
}

func TestFixPreviousLinks_OldNote(t *testing.T) {
	tempDir := t.TempDir()
	journalDir := filepath.Join(tempDir, "journal")

	if err := os.MkdirAll(journalDir, 0755); err != nil {
		t.Fatalf("failed to create journal dir: %v", err)
	}

	// Create a note from 10 days ago
	oldDate := time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC)
	oldJournalPath := filepath.Join(journalDir, oldDate.Format(notes.DateFormat)+".md")
	oldJournalContent := `---
title: Old Journal
---

# Daily Log 2025-01-11

* [Tomorrow](../journal/2025-01-12.md)
`
	if err := os.WriteFile(oldJournalPath, []byte(oldJournalContent), 0644); err != nil {
		t.Fatalf("failed to create old journal: %v", err)
	}

	// Configure
	cfg = &config.Config{
		Journal: config.JournalConfig{
			Dir:            journalDir,
			LinkNextTitles: []string{"Tomorrow"},
		},
		SearchWindowDays: 30,
	}

	// Call fixPreviousLinks for a date that's more than 7 days after the old note
	currentDate := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)
	err := fixPreviousLinks(currentDate, notes.NoteTypeJournal, journalDir)

	// Should not return error
	if err != nil {
		t.Errorf("expected no error for old note, got: %v", err)
	}

	// Verify the old note was NOT modified
	content, err := os.ReadFile(oldJournalPath)
	if err != nil {
		t.Fatalf("failed to read old journal: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "[Tomorrow](../journal/2025-01-12.md)") {
		t.Errorf("expected old note to remain unchanged, got:\n%s", contentStr)
	}
}
