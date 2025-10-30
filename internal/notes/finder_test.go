package notes

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNoteTypeIsValid(t *testing.T) {
	tests := []struct {
		noteType NoteType
		valid    bool
	}{
		{NoteTypeJournal, true},
		{NoteTypeStandup, true},
		{NoteType("invalid"), false},
		{NoteType(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.noteType), func(t *testing.T) {
			if got := tt.noteType.IsValid(); got != tt.valid {
				t.Errorf("IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestParseDateFromFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
		wantErr  bool
	}{
		{
			name:     "valid date",
			filename: "2025-01-06.md",
			want:     "2025-01-06",
			wantErr:  false,
		},
		{
			name:     "valid date with path",
			filename: "/path/to/2025-01-07.md",
			want:     "2025-01-07",
			wantErr:  false,
		},
		{
			name:     "invalid date format",
			filename: "invalid-date.md",
			wantErr:  true,
		},
		{
			name:     "too short",
			filename: "short.md",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDateFromFilename(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDateFromFilename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Format(DateFormat) != tt.want {
				t.Errorf("ParseDateFromFilename() = %v, want %v", got.Format(DateFormat), tt.want)
			}
		})
	}
}

func TestGenerateFilename(t *testing.T) {
	date := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	expected := "2025-01-06.md"

	got := GenerateFilename(date)
	if got != expected {
		t.Errorf("GenerateFilename() = %v, want %v", got, expected)
	}
}

func TestFindNoteByDateExact(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()

	// Create some test files
	testDates := []string{"2025-01-06", "2025-01-07", "2025-01-10"}
	for _, dateStr := range testDates {
		filename := filepath.Join(tmpDir, dateStr+".md")
		if err := os.WriteFile(filename, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	// Test finding exact date
	date := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	path, err := FindNoteByDate(date, NoteTypeJournal, tmpDir, 30)
	if err != nil {
		t.Fatalf("FindNoteByDate() failed: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "2025-01-06.md")
	if path != expectedPath {
		t.Errorf("FindNoteByDate() = %v, want %v", path, expectedPath)
	}
}

func TestFindNoteByDateFallback(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()

	// Create files with gaps (like our testdata)
	// 01-06, 01-07, 01-08 exist, 01-09 missing, 01-10 exists
	testDates := []string{"2025-01-06", "2025-01-07", "2025-01-08", "2025-01-10"}
	for _, dateStr := range testDates {
		filename := filepath.Join(tmpDir, dateStr+".md")
		if err := os.WriteFile(filename, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	// Try to find 01-09 (missing), should fall back to 01-08
	date := time.Date(2025, 1, 9, 0, 0, 0, 0, time.UTC)
	path, err := FindNoteByDate(date, NoteTypeJournal, tmpDir, 30)
	if err != nil {
		t.Fatalf("FindNoteByDate() failed: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "2025-01-08.md")
	if path != expectedPath {
		t.Errorf("FindNoteByDate() = %v, want %v", path, expectedPath)
	}
}

func TestFindNoteByDateMultiDayGap(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()

	// Simulate holiday gap: files on 01-06, then nothing until 01-20
	testDates := []string{"2025-01-06", "2025-01-20"}
	for _, dateStr := range testDates {
		filename := filepath.Join(tmpDir, dateStr+".md")
		if err := os.WriteFile(filename, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	// Try to find 01-15 (middle of gap), should fall back to 01-06
	date := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	path, err := FindNoteByDate(date, NoteTypeJournal, tmpDir, 30)
	if err != nil {
		t.Fatalf("FindNoteByDate() failed: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "2025-01-06.md")
	if path != expectedPath {
		t.Errorf("FindNoteByDate() = %v, want %v", path, expectedPath)
	}
}

func TestFindNoteByDateNotFound(t *testing.T) {
	// Create temp directory with only one file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "2025-01-06.md")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Try to find a date too far in the future (outside search window)
	date := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	_, err := FindNoteByDate(date, NoteTypeJournal, tmpDir, 30)
	if err == nil {
		t.Error("FindNoteByDate() should fail when no note found within window")
	}
}

func TestFindNoteByDateInvalidDirectory(t *testing.T) {
	date := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	_, err := FindNoteByDate(date, NoteTypeJournal, "/nonexistent/directory", 30)
	if err == nil {
		t.Error("FindNoteByDate() should fail for non-existent directory")
	}
}

func TestFindNoteByDateInvalidNoteType(t *testing.T) {
	tmpDir := t.TempDir()
	date := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	_, err := FindNoteByDate(date, NoteType("invalid"), tmpDir, 30)
	if err == nil {
		t.Error("FindNoteByDate() should fail for invalid note type")
	}
}

func TestFindNoteByDateInvalidSearchWindow(t *testing.T) {
	tmpDir := t.TempDir()
	date := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)
	_, err := FindNoteByDate(date, NoteTypeJournal, tmpDir, 0)
	if err == nil {
		t.Error("FindNoteByDate() should fail for zero search window")
	}
}

func TestFindNextNote(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()

	// Create files with gaps
	testDates := []string{"2025-01-06", "2025-01-07", "2025-01-10", "2025-01-13"}
	for _, dateStr := range testDates {
		filename := filepath.Join(tmpDir, dateStr+".md")
		if err := os.WriteFile(filename, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	tests := []struct {
		name      string
		startDate string
		wantDate  string
		wantErr   bool
	}{
		{
			name:      "next day exists",
			startDate: "2025-01-06",
			wantDate:  "2025-01-07",
			wantErr:   false,
		},
		{
			name:      "skip gap to find next",
			startDate: "2025-01-07",
			wantDate:  "2025-01-10",
			wantErr:   false,
		},
		{
			name:      "skip weekend gap",
			startDate: "2025-01-10",
			wantDate:  "2025-01-13",
			wantErr:   false,
		},
		{
			name:      "no next note",
			startDate: "2025-01-13",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, _ := time.Parse(DateFormat, tt.startDate)
			path, err := FindNextNote(date, NoteTypeJournal, tmpDir, 30)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindNextNote() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				expectedPath := filepath.Join(tmpDir, tt.wantDate+".md")
				if path != expectedPath {
					t.Errorf("FindNextNote() = %v, want %v", path, expectedPath)
				}
			}
		})
	}
}

func TestFindNextNoteInvalidInputs(t *testing.T) {
	tmpDir := t.TempDir()
	date := time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)

	// Invalid note type
	_, err := FindNextNote(date, NoteType("invalid"), tmpDir, 30)
	if err == nil {
		t.Error("FindNextNote() should fail for invalid note type")
	}

	// Invalid search window
	_, err = FindNextNote(date, NoteTypeJournal, tmpDir, -1)
	if err == nil {
		t.Error("FindNextNote() should fail for negative search window")
	}
}

// TestWithRealTestData tests finder functions with actual testdata
func TestWithRealTestData(t *testing.T) {
	// Test with real testdata directory
	testdataDir := "../../testdata/journal"

	// Check if testdata exists
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		t.Skip("Testdata directory not found, skipping integration test")
	}

	tests := []struct {
		name        string
		searchDate  string
		expectFound bool
		expectDate  string // Date we expect to find (if different from search)
	}{
		{
			name:        "exact match 2025-01-06",
			searchDate:  "2025-01-06",
			expectFound: true,
			expectDate:  "2025-01-06",
		},
		{
			name:        "fallback from missing 2025-01-09 to 2025-01-08",
			searchDate:  "2025-01-09",
			expectFound: true,
			expectDate:  "2025-01-08",
		},
		{
			name:        "weekend fallback from 2025-01-11 to 2025-01-10",
			searchDate:  "2025-01-11",
			expectFound: true,
			expectDate:  "2025-01-10",
		},
		{
			name:        "exact match after gap 2025-01-13",
			searchDate:  "2025-01-13",
			expectFound: true,
			expectDate:  "2025-01-13",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, _ := time.Parse(DateFormat, tt.searchDate)
			path, err := FindNoteByDate(date, NoteTypeJournal, testdataDir, 30)

			if tt.expectFound && err != nil {
				t.Errorf("FindNoteByDate() failed: %v", err)
				return
			}

			if tt.expectFound {
				expectedFilename := tt.expectDate + ".md"
				if filepath.Base(path) != expectedFilename {
					t.Errorf("FindNoteByDate() found %s, want %s", filepath.Base(path), expectedFilename)
				}
				t.Logf("Found: %s", path)
			}
		})
	}
}

// TestFindNextNoteWithRealTestData tests FindNextNote with real testdata
func TestFindNextNoteWithRealTestData(t *testing.T) {
	testdataDir := "../../testdata/journal"

	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		t.Skip("Testdata directory not found, skipping integration test")
	}

	// Test finding next note after 2025-01-08 (should skip 01-09 and find 01-10)
	date, _ := time.Parse(DateFormat, "2025-01-08")
	path, err := FindNextNote(date, NoteTypeJournal, testdataDir, 30)
	if err != nil {
		t.Fatalf("FindNextNote() failed: %v", err)
	}

	if filepath.Base(path) != "2025-01-10.md" {
		t.Errorf("FindNextNote() = %s, want 2025-01-10.md", filepath.Base(path))
	}

	t.Logf("Successfully found next note: %s", path)
}
