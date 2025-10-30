package notes

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// DateFormat is the format used for note filenames (YYYY-MM-DD)
	DateFormat = "2006-01-02"
)

// FindNoteByDate finds a note file for the given date, with fallback to previous dates
// within the search window if the exact date doesn't exist.
//
// Parameters:
//   - date: the target date to find
//   - noteType: the type of note (journal or standup)
//   - dir: the directory to search in
//   - searchWindowDays: how many days back to search if exact date not found
//
// Returns:
//   - the absolute path to the found note file
//   - error if no note found within search window or other errors
func FindNoteByDate(date time.Time, noteType NoteType, dir string, searchWindowDays int) (string, error) {
	if !noteType.IsValid() {
		return "", fmt.Errorf("invalid note type: %s", noteType)
	}

	if searchWindowDays <= 0 {
		return "", fmt.Errorf("searchWindowDays must be positive, got %d", searchWindowDays)
	}

	// Ensure directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", fmt.Errorf("directory does not exist: %s", dir)
	}

	// Try exact date first
	exactPath := filepath.Join(dir, date.Format(DateFormat)+".md")
	if fileExists(exactPath) {
		return exactPath, nil
	}

	// Fall back to searching previous dates within window
	for i := 1; i <= searchWindowDays; i++ {
		previousDate := date.AddDate(0, 0, -i)
		previousPath := filepath.Join(dir, previousDate.Format(DateFormat)+".md")

		if fileExists(previousPath) {
			return previousPath, nil
		}
	}

	// No note found within search window
	return "", fmt.Errorf(
		"no %s note found for %s or within %d days before",
		noteType,
		date.Format(DateFormat),
		searchWindowDays,
	)
}

// FindNextNote finds the next note file after the given date
// within the search window.
//
// Parameters:
//   - date: the starting date
//   - noteType: the type of note (journal or standup)
//   - dir: the directory to search in
//   - searchWindowDays: how many days forward to search
//
// Returns:
//   - the absolute path to the found note file
//   - error if no note found within search window
func FindNextNote(date time.Time, noteType NoteType, dir string, searchWindowDays int) (string, error) {
	if !noteType.IsValid() {
		return "", fmt.Errorf("invalid note type: %s", noteType)
	}

	if searchWindowDays <= 0 {
		return "", fmt.Errorf("searchWindowDays must be positive, got %d", searchWindowDays)
	}

	// Search forward from the next day
	for i := 1; i <= searchWindowDays; i++ {
		nextDate := date.AddDate(0, 0, i)
		nextPath := filepath.Join(dir, nextDate.Format(DateFormat)+".md")

		if fileExists(nextPath) {
			return nextPath, nil
		}
	}

	// No note found within search window
	return "", fmt.Errorf(
		"no %s note found after %s within %d days",
		noteType,
		date.Format(DateFormat),
		searchWindowDays,
	)
}

// ParseDateFromFilename extracts the date from a note filename
// Expected format: YYYY-MM-DD.md
func ParseDateFromFilename(filename string) (time.Time, error) {
	// Remove extension
	base := filepath.Base(filename)
	if len(base) < 10 {
		return time.Time{}, fmt.Errorf("filename too short: %s", filename)
	}

	// Try to parse the date part (first 10 characters)
	dateStr := base[:10]
	date, err := time.Parse(DateFormat, dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format in filename %s: %w", filename, err)
	}

	return date, nil
}

// GenerateFilename generates a filename for a note of the given date
func GenerateFilename(date time.Time) string {
	return date.Format(DateFormat) + ".md"
}

// fileExists checks if a file exists and is not a directory
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
