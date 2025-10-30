package links

import (
	"testing"
	"time"

	"github.com/rdark/za/internal/config"
	"github.com/rdark/za/internal/markdown"
	"github.com/rdark/za/internal/notes"
)

func TestResolvePreviousLink(t *testing.T) {
	// Use testdata config
	cfg := config.DefaultConfig()
	cfg.Journal.Dir = "../../testdata/journal"
	cfg.Standup.Dir = "../../testdata/standup"

	// Current date: 2025-01-07 (Tuesday)
	currentDate := time.Date(2025, 1, 7, 0, 0, 0, 0, time.UTC)
	resolver := NewResolver(cfg, currentDate, notes.NoteTypeJournal)

	// Classify a "Yesterday" link
	link := markdown.Link{
		Text:        "Yesterday",
		Destination: "2025-01-06", // Currently points to 2025-01-06
	}

	classifier := NewClassifier(cfg)
	classified := classifier.Classify(link)

	// Resolve it
	resolved := resolver.Resolve(classified)

	if resolved.Error != nil {
		t.Fatalf("Resolve() error = %v", resolved.Error)
	}

	// Should find 2025-01-06 (the day before)
	expectedDate := "2025-01-06"
	if resolved.ResolvedDate.Format(notes.DateFormat) != expectedDate {
		t.Errorf("ResolvedDate = %v, want %v", resolved.ResolvedDate.Format(notes.DateFormat), expectedDate)
	}

	// Link is already correct, shouldn't need update
	if resolved.NeedsUpdate {
		t.Error("Link shouldn't need update when already correct")
	}

	t.Logf("Resolved: %s -> %s", link.Destination, resolved.ResolvedDate.Format(notes.DateFormat))
}

func TestResolveNextLink(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Journal.Dir = "../../testdata/journal"
	cfg.Standup.Dir = "../../testdata/standup"

	// Current date: 2025-01-08 (Wednesday)
	currentDate := time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC)
	resolver := NewResolver(cfg, currentDate, notes.NoteTypeJournal)

	// Classify a "Tomorrow" link
	link := markdown.Link{
		Text:        "Tomorrow",
		Destination: "2025-01-09", // Points to missing day
	}

	classifier := NewClassifier(cfg)
	classified := classifier.Classify(link)

	// Resolve it
	resolved := resolver.Resolve(classified)

	if resolved.Error != nil {
		t.Fatalf("Resolve() error = %v", resolved.Error)
	}

	// Should find 2025-01-10 (skipping missing 01-09)
	expectedDate := "2025-01-10"
	if resolved.ResolvedDate.Format(notes.DateFormat) != expectedDate {
		t.Errorf("ResolvedDate = %v, want %v", resolved.ResolvedDate.Format(notes.DateFormat), expectedDate)
	}

	// Should need update (01-09 -> 01-10)
	if !resolved.NeedsUpdate {
		t.Error("Link should need update when pointing to missing day")
	}

	// Check suggested destination
	if resolved.SuggestedDestination != expectedDate {
		t.Errorf("SuggestedDestination = %v, want %v", resolved.SuggestedDestination, expectedDate)
	}

	t.Logf("Resolved: %s -> %s (needs update: %v)", link.Destination, resolved.ResolvedDate.Format(notes.DateFormat), resolved.NeedsUpdate)
}

func TestResolveCrossReference(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Journal.Dir = "../../testdata/journal"
	cfg.Standup.Dir = "../../testdata/standup"

	// Current date: 2025-01-07, in a journal
	currentDate := time.Date(2025, 1, 7, 0, 0, 0, 0, time.UTC)
	resolver := NewResolver(cfg, currentDate, notes.NoteTypeJournal)

	// Cross-reference to standup
	link := markdown.Link{
		Text:        "Standup",
		Destination: "../standup/2025-01-07",
	}

	classifier := NewClassifier(cfg)
	classified := classifier.Classify(link)

	// Resolve it
	resolved := resolver.Resolve(classified)

	if resolved.Error != nil {
		t.Fatalf("Resolve() error = %v", resolved.Error)
	}

	// Should find standup for same date
	expectedDate := "2025-01-07"
	if resolved.ResolvedDate.Format(notes.DateFormat) != expectedDate {
		t.Errorf("ResolvedDate = %v, want %v", resolved.ResolvedDate.Format(notes.DateFormat), expectedDate)
	}

	t.Logf("Resolved cross-reference: %s", resolved.ResolvedPath)
}

func TestResolveWeekendGap(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Journal.Dir = "../../testdata/journal"
	cfg.Standup.Dir = "../../testdata/standup"

	// Current date: 2025-01-13 (Monday after weekend)
	currentDate := time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC)
	resolver := NewResolver(cfg, currentDate, notes.NoteTypeJournal)

	// "Yesterday" link pointing to Sunday
	link := markdown.Link{
		Text:        "Yesterday",
		Destination: "2025-01-12", // Sunday (missing)
	}

	classifier := NewClassifier(cfg)
	classified := classifier.Classify(link)

	// Resolve it
	resolved := resolver.Resolve(classified)

	if resolved.Error != nil {
		t.Fatalf("Resolve() error = %v", resolved.Error)
	}

	// Should find 2025-01-10 (Friday, skipping weekend)
	expectedDate := "2025-01-10"
	if resolved.ResolvedDate.Format(notes.DateFormat) != expectedDate {
		t.Errorf("ResolvedDate = %v, want %v (should skip weekend)", resolved.ResolvedDate.Format(notes.DateFormat), expectedDate)
	}

	// Should need update
	if !resolved.NeedsUpdate {
		t.Error("Link should need update when pointing to missing weekend")
	}

	t.Logf("Resolved weekend gap: %s -> %s", link.Destination, resolved.ResolvedDate.Format(notes.DateFormat))
}

func TestResolveAll(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Journal.Dir = "../../testdata/journal"
	cfg.Standup.Dir = "../../testdata/standup"

	currentDate := time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC)
	resolver := NewResolver(cfg, currentDate, notes.NoteTypeJournal)
	classifier := NewClassifier(cfg)

	// Multiple links
	links := []markdown.Link{
		{Text: "Yesterday", Destination: "2025-01-07"},
		{Text: "Tomorrow", Destination: "2025-01-09"},
		{Text: "Standup", Destination: "../standup/2025-01-08"},
		{Text: "External", Destination: "https://example.com"},
	}

	// Classify all
	classified := classifier.ClassifyAll(links)

	// Resolve all
	resolved := resolver.ResolveAll(classified)

	if len(resolved) != 4 {
		t.Fatalf("ResolveAll() returned %d links, want 4", len(resolved))
	}

	// Check that some were resolved
	resolvedCount := 0
	for _, r := range resolved {
		if r.ResolvedPath != "" {
			resolvedCount++
			t.Logf("Resolved: [%s](%s) -> %s", r.Classified.Link.Text, r.Classified.Link.Destination, r.ResolvedDate.Format(notes.DateFormat))
		}
	}

	if resolvedCount == 0 {
		t.Error("Expected some links to be resolved")
	}
}

func TestFilterNeedsUpdate(t *testing.T) {
	resolved := []ResolvedLink{
		{NeedsUpdate: true, ResolvedDate: time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC)},
		{NeedsUpdate: false, ResolvedDate: time.Date(2025, 1, 7, 0, 0, 0, 0, time.UTC)},
		{NeedsUpdate: true, ResolvedDate: time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC)},
		{NeedsUpdate: false, ResolvedDate: time.Date(2025, 1, 9, 0, 0, 0, 0, time.UTC)},
	}

	needsUpdate := FilterNeedsUpdate(resolved)

	if len(needsUpdate) != 2 {
		t.Errorf("FilterNeedsUpdate() = %d links, want 2", len(needsUpdate))
	}

	for _, r := range needsUpdate {
		if !r.NeedsUpdate {
			t.Error("Filtered link should have NeedsUpdate = true")
		}
	}
}

func TestResolveWithDifferentNoteTypes(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Journal.Dir = "../../testdata/journal"
	cfg.Standup.Dir = "../../testdata/standup"

	currentDate := time.Date(2025, 1, 7, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		currentNoteType notes.NoteType
		linkText        string
		linkDest        string
		expectResolved  bool
	}{
		{
			name:            "journal yesterday",
			currentNoteType: notes.NoteTypeJournal,
			linkText:        "Yesterday",
			linkDest:        "2025-01-06",
			expectResolved:  true,
		},
		{
			name:            "standup yesterday",
			currentNoteType: notes.NoteTypeStandup,
			linkText:        "Yesterday",
			linkDest:        "2025-01-06",
			expectResolved:  true,
		},
		{
			name:            "journal to standup cross-ref",
			currentNoteType: notes.NoteTypeJournal,
			linkText:        "Standup",
			linkDest:        "../standup/2025-01-07",
			expectResolved:  true,
		},
		{
			name:            "standup to journal cross-ref",
			currentNoteType: notes.NoteTypeStandup,
			linkText:        "Daily",
			linkDest:        "../journal/2025-01-07",
			expectResolved:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewResolver(cfg, currentDate, tt.currentNoteType)
			classifier := NewClassifier(cfg)

			link := markdown.Link{
				Text:        tt.linkText,
				Destination: tt.linkDest,
			}

			classified := classifier.Classify(link)
			resolved := resolver.Resolve(classified)

			if tt.expectResolved {
				if resolved.Error != nil {
					t.Errorf("Expected link to resolve, got error: %v", resolved.Error)
				}
				if resolved.ResolvedPath == "" {
					t.Error("Expected ResolvedPath to be set")
				}
				t.Logf("Resolved: %s", resolved.ResolvedPath)
			}
		})
	}
}
