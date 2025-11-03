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

func TestFixCrossReferenceLinks_StandupFixesJournal(t *testing.T) {
	tempDir := t.TempDir()
	journalDir := filepath.Join(tempDir, "journal")
	standupDir := filepath.Join(tempDir, "standup")

	if err := os.MkdirAll(journalDir, 0755); err != nil {
		t.Fatalf("failed to create journal dir: %v", err)
	}
	if err := os.MkdirAll(standupDir, 0755); err != nil {
		t.Fatalf("failed to create standup dir: %v", err)
	}

	// Create today's journal with standup link pointing to yesterday
	currentDate := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)
	journalPath := filepath.Join(journalDir, currentDate.Format(notes.DateFormat)+".md")
	journalContent := `---
title: Journal 2025-01-21
---

# Daily Log 2025-01-21

* [Yesterday](../journal/2025-01-20.md)
* [Tomorrow](../journal/2025-01-22.md)
* [Standup](../standup/2025-01-20.md)

## Goals of the Day

* [ ] Test goal
`
	if err := os.WriteFile(journalPath, []byte(journalContent), 0644); err != nil {
		t.Fatalf("failed to create journal: %v", err)
	}

	// Configure
	cfg = &config.Config{
		Journal: config.JournalConfig{
			Dir: journalDir,
		},
		Standup: config.StandupConfig{
			Dir: standupDir,
		},
		SearchWindowDays: 30,
	}

	// Call fixCrossReferenceLinks to fix journal's standup link
	err := fixCrossReferenceLinks(currentDate, notes.NoteTypeJournal, notes.NoteTypeStandup, journalDir)
	if err != nil {
		t.Fatalf("fixCrossReferenceLinks failed: %v", err)
	}

	// Read the updated journal
	updatedContent, err := os.ReadFile(journalPath)
	if err != nil {
		t.Fatalf("failed to read updated journal: %v", err)
	}

	contentStr := string(updatedContent)

	// Verify the "Standup" link was updated to point to 2025-01-21
	if !strings.Contains(contentStr, "[Standup](../standup/2025-01-21.md)") {
		t.Errorf("expected Standup link to be updated to 2025-01-21.md, got:\n%s", contentStr)
	}

	// Verify other links were not modified
	if !strings.Contains(contentStr, "[Yesterday](../journal/2025-01-20.md)") {
		t.Errorf("expected Yesterday link to remain unchanged, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "[Tomorrow](../journal/2025-01-22.md)") {
		t.Errorf("expected Tomorrow link to remain unchanged, got:\n%s", contentStr)
	}
}

func TestFixCrossReferenceLinks_JournalFixesStandup(t *testing.T) {
	tempDir := t.TempDir()
	journalDir := filepath.Join(tempDir, "journal")
	standupDir := filepath.Join(tempDir, "standup")

	if err := os.MkdirAll(journalDir, 0755); err != nil {
		t.Fatalf("failed to create journal dir: %v", err)
	}
	if err := os.MkdirAll(standupDir, 0755); err != nil {
		t.Fatalf("failed to create standup dir: %v", err)
	}

	// Create today's standup with journal link pointing to yesterday
	currentDate := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)
	standupPath := filepath.Join(standupDir, currentDate.Format(notes.DateFormat)+".md")
	standupContent := `---
title: Standup 2025-01-21
---

# Standup 2025-01-21

* [Previous](../standup/2025-01-20.md)
* [Next](../standup/2025-01-22.md)
* [Journal](../journal/2025-01-20.md)

## Worked on Yesterday

* Some work

## Working on Today

* Some goals
`
	if err := os.WriteFile(standupPath, []byte(standupContent), 0644); err != nil {
		t.Fatalf("failed to create standup: %v", err)
	}

	// Configure
	cfg = &config.Config{
		Journal: config.JournalConfig{
			Dir: journalDir,
		},
		Standup: config.StandupConfig{
			Dir: standupDir,
		},
		SearchWindowDays: 30,
	}

	// Call fixCrossReferenceLinks to fix standup's journal link
	err := fixCrossReferenceLinks(currentDate, notes.NoteTypeStandup, notes.NoteTypeJournal, standupDir)
	if err != nil {
		t.Fatalf("fixCrossReferenceLinks failed: %v", err)
	}

	// Read the updated standup
	updatedContent, err := os.ReadFile(standupPath)
	if err != nil {
		t.Fatalf("failed to read updated standup: %v", err)
	}

	contentStr := string(updatedContent)

	// Verify the "Journal" link was updated to point to 2025-01-21
	if !strings.Contains(contentStr, "[Journal](../journal/2025-01-21.md)") {
		t.Errorf("expected Journal link to be updated to 2025-01-21.md, got:\n%s", contentStr)
	}

	// Verify other links were not modified
	if !strings.Contains(contentStr, "[Previous](../standup/2025-01-20.md)") {
		t.Errorf("expected Previous link to remain unchanged, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "[Next](../standup/2025-01-22.md)") {
		t.Errorf("expected Next link to remain unchanged, got:\n%s", contentStr)
	}
}

func TestFixCrossReferenceLinks_NoTargetNote(t *testing.T) {
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

	// Call fixCrossReferenceLinks when no target note exists
	currentDate := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)
	err := fixCrossReferenceLinks(currentDate, notes.NoteTypeJournal, notes.NoteTypeStandup, journalDir)

	// Should not return error when no target note exists
	if err != nil {
		t.Errorf("expected no error when no target note exists, got: %v", err)
	}
}

func TestFixCrossReferenceLinks_LinkAlreadyCorrect(t *testing.T) {
	tempDir := t.TempDir()
	journalDir := filepath.Join(tempDir, "journal")
	standupDir := filepath.Join(tempDir, "standup")

	if err := os.MkdirAll(journalDir, 0755); err != nil {
		t.Fatalf("failed to create journal dir: %v", err)
	}
	if err := os.MkdirAll(standupDir, 0755); err != nil {
		t.Fatalf("failed to create standup dir: %v", err)
	}

	// Create today's journal with standup link already pointing to today
	currentDate := time.Date(2025, 1, 21, 0, 0, 0, 0, time.UTC)
	journalPath := filepath.Join(journalDir, currentDate.Format(notes.DateFormat)+".md")
	journalContent := `---
title: Journal 2025-01-21
---

# Daily Log 2025-01-21

* [Standup](../standup/2025-01-21.md)

## Goals of the Day

* [ ] Test goal
`
	if err := os.WriteFile(journalPath, []byte(journalContent), 0644); err != nil {
		t.Fatalf("failed to create journal: %v", err)
	}

	// Configure
	cfg = &config.Config{
		Journal: config.JournalConfig{
			Dir: journalDir,
		},
		Standup: config.StandupConfig{
			Dir: standupDir,
		},
		SearchWindowDays: 30,
	}

	// Call fixCrossReferenceLinks - should not modify anything
	err := fixCrossReferenceLinks(currentDate, notes.NoteTypeJournal, notes.NoteTypeStandup, journalDir)
	if err != nil {
		t.Fatalf("fixCrossReferenceLinks failed: %v", err)
	}

	// Read the journal
	content, err := os.ReadFile(journalPath)
	if err != nil {
		t.Fatalf("failed to read journal: %v", err)
	}

	contentStr := string(content)

	// Verify link is still correct
	if !strings.Contains(contentStr, "[Standup](../standup/2025-01-21.md)") {
		t.Errorf("expected Standup link to remain as 2025-01-21.md, got:\n%s", contentStr)
	}
}
