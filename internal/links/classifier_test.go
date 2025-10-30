package links

import (
	"testing"

	"github.com/rdark/za/internal/config"
	"github.com/rdark/za/internal/markdown"
)

func TestClassify(t *testing.T) {
	cfg := config.DefaultConfig()
	classifier := NewClassifier(cfg)

	tests := []struct {
		name           string
		link           markdown.Link
		expectedType   LinkType
		expectedTarget string
	}{
		{
			name: "yesterday link",
			link: markdown.Link{
				Text:        "Yesterday",
				Destination: "2025-01-05",
			},
			expectedType: LinkTypeTemporalPrevious,
		},
		{
			name: "previous link (synonym)",
			link: markdown.Link{
				Text:        "Previous",
				Destination: "2025-01-05",
			},
			expectedType: LinkTypeTemporalPrevious,
		},
		{
			name: "tomorrow link",
			link: markdown.Link{
				Text:        "Tomorrow",
				Destination: "2025-01-07",
			},
			expectedType: LinkTypeTemporalNext,
		},
		{
			name: "next link (synonym)",
			link: markdown.Link{
				Text:        "Next",
				Destination: "2025-01-07",
			},
			expectedType: LinkTypeTemporalNext,
		},
		{
			name: "standup cross-reference",
			link: markdown.Link{
				Text:        "Standup",
				Destination: "../standup/2025-01-06.md",
			},
			expectedType:   LinkTypeCrossReference,
			expectedTarget: "standup",
		},
		{
			name: "journal cross-reference",
			link: markdown.Link{
				Text:        "Daily",
				Destination: "../journal/2025-01-06.md",
			},
			expectedType:   LinkTypeCrossReference,
			expectedTarget: "journal",
		},
		{
			name: "external link",
			link: markdown.Link{
				Text:        "External",
				Destination: "https://example.com",
			},
			expectedType: LinkTypeExternal,
		},
		{
			name: "wiki link",
			link: markdown.Link{
				Text:        "Some Page",
				Destination: "some-page",
			},
			expectedType: LinkTypeOther,
		},
		{
			name: "case insensitive yesterday",
			link: markdown.Link{
				Text:        "YESTERDAY",
				Destination: "2025-01-05",
			},
			expectedType: LinkTypeTemporalPrevious,
		},
		{
			name: "last week synonym",
			link: markdown.Link{
				Text:        "Last Week",
				Destination: "2024-12-30",
			},
			expectedType: LinkTypeTemporalPrevious,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classified := classifier.Classify(tt.link)

			if classified.Type != tt.expectedType {
				t.Errorf("Classify() type = %v, want %v", classified.Type, tt.expectedType)
			}

			if tt.expectedTarget != "" && classified.TargetNoteType != tt.expectedTarget {
				t.Errorf("Classify() target = %v, want %v", classified.TargetNoteType, tt.expectedTarget)
			}
		})
	}
}

func TestClassifyAll(t *testing.T) {
	cfg := config.DefaultConfig()
	classifier := NewClassifier(cfg)

	links := []markdown.Link{
		{Text: "Yesterday", Destination: "2025-01-05"},
		{Text: "Tomorrow", Destination: "2025-01-07"},
		{Text: "External", Destination: "https://example.com"},
		{Text: "Standup", Destination: "../standup/2025-01-06.md"},
	}

	classified := classifier.ClassifyAll(links)

	if len(classified) != 4 {
		t.Errorf("ClassifyAll() returned %d links, want 4", len(classified))
	}

	// Check types
	expectedTypes := []LinkType{
		LinkTypeTemporalPrevious,
		LinkTypeTemporalNext,
		LinkTypeExternal,
		LinkTypeCrossReference,
	}

	for i, expected := range expectedTypes {
		if classified[i].Type != expected {
			t.Errorf("Link %d: type = %v, want %v", i, classified[i].Type, expected)
		}
	}
}

func TestFilterByType(t *testing.T) {
	links := []ClassifiedLink{
		{Link: markdown.Link{Text: "Yesterday"}, Type: LinkTypeTemporalPrevious},
		{Link: markdown.Link{Text: "Tomorrow"}, Type: LinkTypeTemporalNext},
		{Link: markdown.Link{Text: "External"}, Type: LinkTypeExternal},
		{Link: markdown.Link{Text: "Standup"}, Type: LinkTypeCrossReference},
		{Link: markdown.Link{Text: "Yesterday2"}, Type: LinkTypeTemporalPrevious},
	}

	// Filter temporal previous
	previous := FilterByType(links, LinkTypeTemporalPrevious)
	if len(previous) != 2 {
		t.Errorf("FilterByType(temporal_previous) = %d links, want 2", len(previous))
	}

	// Filter external
	external := FilterByType(links, LinkTypeExternal)
	if len(external) != 1 {
		t.Errorf("FilterByType(external) = %d links, want 1", len(external))
	}

	// Filter cross-reference
	crossRef := FilterByType(links, LinkTypeCrossReference)
	if len(crossRef) != 1 {
		t.Errorf("FilterByType(cross_reference) = %d links, want 1", len(crossRef))
	}
}

func TestNeedsFixing(t *testing.T) {
	tests := []struct {
		name string
		link ClassifiedLink
		want bool
	}{
		{
			name: "temporal previous needs fixing",
			link: ClassifiedLink{
				Link: markdown.Link{Destination: "2025-01-05"},
				Type: LinkTypeTemporalPrevious,
			},
			want: true,
		},
		{
			name: "temporal next needs fixing",
			link: ClassifiedLink{
				Link: markdown.Link{Destination: "2025-01-07"},
				Type: LinkTypeTemporalNext,
			},
			want: true,
		},
		{
			name: "cross-reference needs fixing",
			link: ClassifiedLink{
				Link: markdown.Link{Destination: "../standup/2025-01-06.md"},
				Type: LinkTypeCrossReference,
			},
			want: true,
		},
		{
			name: "external doesn't need fixing",
			link: ClassifiedLink{
				Link: markdown.Link{Destination: "https://example.com"},
				Type: LinkTypeExternal,
			},
			want: false,
		},
		{
			name: "other doesn't need fixing",
			link: ClassifiedLink{
				Link: markdown.Link{Destination: "some-page"},
				Type: LinkTypeOther,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.link.NeedsFixing(); got != tt.want {
				t.Errorf("NeedsFixing() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClassifyWithCustomConfig(t *testing.T) {
	// Custom config with additional synonyms
	cfg := config.DefaultConfig()
	cfg.Journal.LinkPreviousTitles = append(cfg.Journal.LinkPreviousTitles, "Before")
	cfg.Journal.LinkNextTitles = append(cfg.Journal.LinkNextTitles, "After")

	classifier := NewClassifier(cfg)

	tests := []struct {
		name string
		link markdown.Link
		want LinkType
	}{
		{
			name: "custom previous synonym",
			link: markdown.Link{
				Text:        "Before",
				Destination: "2025-01-05",
			},
			want: LinkTypeTemporalPrevious,
		},
		{
			name: "custom next synonym",
			link: markdown.Link{
				Text:        "After",
				Destination: "2025-01-07",
			},
			want: LinkTypeTemporalNext,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classified := classifier.Classify(tt.link)
			if classified.Type != tt.want {
				t.Errorf("Classify() = %v, want %v", classified.Type, tt.want)
			}
		})
	}
}

func TestCrossReferencePatterns(t *testing.T) {
	cfg := config.DefaultConfig()
	classifier := NewClassifier(cfg)

	tests := []struct {
		name     string
		linkText string
		want     LinkType
	}{
		{
			name:     "standup reference",
			linkText: "Standup",
			want:     LinkTypeCrossReference,
		},
		{
			name:     "journal reference",
			linkText: "Journal",
			want:     LinkTypeCrossReference,
		},
		{
			name:     "daily reference",
			linkText: "Daily",
			want:     LinkTypeCrossReference,
		},
		{
			name:     "daily log reference",
			linkText: "Daily Log",
			want:     LinkTypeCrossReference,
		},
		{
			name:     "standup yesterday combined",
			linkText: "Standup/Yesterday",
			want:     LinkTypeCrossReference,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link := markdown.Link{
				Text:        tt.linkText,
				Destination: "../standup/2025-01-06.md",
			}
			classified := classifier.Classify(link)
			if classified.Type != tt.want {
				t.Errorf("Classify() = %v, want %v", classified.Type, tt.want)
			}
		})
	}
}
