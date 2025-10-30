package markdown

import (
	"testing"
)

func TestExtractLinks(t *testing.T) {
	content := `---
title: test
---

# Section 1

This has a [link to yesterday](2025-01-05) and [tomorrow](2025-01-07).

Also an [external link](https://example.com) and a [relative link](../standup/2025-01-06).

# Links

* [Yesterday](2025-01-05)
* [Tomorrow](2025-01-07)
* [Standup](../standup/2025-01-06)
`

	p := NewParser()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	links := doc.ExtractLinks()

	if len(links) == 0 {
		t.Fatal("expected to find links")
	}

	// Should find all links
	expectedCount := 7
	if len(links) != expectedCount {
		t.Errorf("expected %d links, got %d", expectedCount, len(links))
	}

	// Check first link
	if links[0].Text != "link to yesterday" {
		t.Errorf("expected text 'link to yesterday', got %q", links[0].Text)
	}
	if links[0].Destination != "2025-01-05" {
		t.Errorf("expected destination '2025-01-05', got %q", links[0].Destination)
	}

	t.Logf("Found %d links:", len(links))
	for i, link := range links {
		t.Logf("  %d: [%s](%s) at line %d", i, link.Text, link.Destination, link.Line)
	}
}

func TestIsDateLink(t *testing.T) {
	tests := []struct {
		name        string
		destination string
		want        bool
	}{
		{
			name:        "simple date",
			destination: "2025-01-06",
			want:        true,
		},
		{
			name:        "date with .md",
			destination: "2025-01-06.md",
			want:        true,
		},
		{
			name:        "relative journal path",
			destination: "../journal/2025-01-06.md",
			want:        true,
		},
		{
			name:        "relative standup path",
			destination: "../standup/2025-01-07",
			want:        true,
		},
		{
			name:        "external URL",
			destination: "https://example.com",
			want:        false,
		},
		{
			name:        "wiki link",
			destination: "some-page",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link := Link{Destination: tt.destination}
			if got := link.IsDateLink(); got != tt.want {
				t.Errorf("IsDateLink() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRelativeLink(t *testing.T) {
	tests := []struct {
		name        string
		destination string
		want        bool
	}{
		{
			name:        "relative with ../",
			destination: "../journal/2025-01-06.md",
			want:        true,
		},
		{
			name:        "relative with ./",
			destination: "./2025-01-06.md",
			want:        true,
		},
		{
			name:        "simple relative",
			destination: "2025-01-06",
			want:        true,
		},
		{
			name:        "external URL",
			destination: "https://example.com",
			want:        false,
		},
		{
			name:        "http URL",
			destination: "http://example.com",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link := Link{Destination: tt.destination}
			if got := link.IsRelativeLink(); got != tt.want {
				t.Errorf("IsRelativeLink() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsExternalLink(t *testing.T) {
	tests := []struct {
		name        string
		destination string
		want        bool
	}{
		{
			name:        "https URL",
			destination: "https://example.com",
			want:        true,
		},
		{
			name:        "http URL",
			destination: "http://example.com",
			want:        true,
		},
		{
			name:        "relative path",
			destination: "../journal/2025-01-06.md",
			want:        false,
		},
		{
			name:        "date link",
			destination: "2025-01-06",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link := Link{Destination: tt.destination}
			if got := link.IsExternalLink(); got != tt.want {
				t.Errorf("IsExternalLink() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDateFromDestination(t *testing.T) {
	tests := []struct {
		name        string
		destination string
		want        string
	}{
		{
			name:        "simple date",
			destination: "2025-01-06",
			want:        "2025-01-06",
		},
		{
			name:        "date with .md",
			destination: "2025-01-06.md",
			want:        "2025-01-06",
		},
		{
			name:        "relative path with date",
			destination: "../journal/2025-01-06.md",
			want:        "2025-01-06",
		},
		{
			name:        "no date",
			destination: "some-page",
			want:        "",
		},
		{
			name:        "URL with no date",
			destination: "https://example.com",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link := Link{Destination: tt.destination}
			if got := link.GetDateFromDestination(); got != tt.want {
				t.Errorf("GetDateFromDestination() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetNoteTypeFromDestination(t *testing.T) {
	tests := []struct {
		name        string
		destination string
		want        string
	}{
		{
			name:        "journal path",
			destination: "../journal/2025-01-06.md",
			want:        "journal",
		},
		{
			name:        "standup path",
			destination: "../standup/2025-01-06.md",
			want:        "standup",
		},
		{
			name:        "journal without ../",
			destination: "journal/2025-01-06",
			want:        "journal",
		},
		{
			name:        "case insensitive",
			destination: "../Journal/2025-01-06",
			want:        "journal",
		},
		{
			name:        "unknown type",
			destination: "2025-01-06",
			want:        "",
		},
		{
			name:        "external link",
			destination: "https://example.com",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link := Link{Destination: tt.destination}
			if got := link.GetNoteTypeFromDestination(); got != tt.want {
				t.Errorf("GetNoteTypeFromDestination() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFilterLinks(t *testing.T) {
	links := []Link{
		{Text: "date1", Destination: "2025-01-06"},
		{Text: "date2", Destination: "../journal/2025-01-07.md"},
		{Text: "external", Destination: "https://example.com"},
		{Text: "wiki", Destination: "some-page"},
	}

	// Filter to only date links
	dateLinks := FilterLinks(links, func(l Link) bool {
		return l.IsDateLink()
	})

	if len(dateLinks) != 2 {
		t.Errorf("expected 2 date links, got %d", len(dateLinks))
	}

	// Filter to only external links
	externalLinks := FilterLinks(links, func(l Link) bool {
		return l.IsExternalLink()
	})

	if len(externalLinks) != 1 {
		t.Errorf("expected 1 external link, got %d", len(externalLinks))
	}
}

func TestExtractLinksRealTestData(t *testing.T) {
	testFile := "../../testdata/journal/2025-01-06.md"

	p := NewParser()
	doc, err := p.ParseFile(testFile)
	if err != nil {
		t.Skipf("Skipping, testdata file not accessible: %v", err)
		return
	}

	links := doc.ExtractLinks()

	if len(links) == 0 {
		t.Fatal("expected to find links in testdata")
	}

	t.Logf("Found %d links in testdata file", len(links))

	// Count different types of links
	dateLinks := FilterLinks(links, func(l Link) bool { return l.IsDateLink() })
	externalLinks := FilterLinks(links, func(l Link) bool { return l.IsExternalLink() })

	t.Logf("  Date links: %d", len(dateLinks))
	t.Logf("  External links: %d", len(externalLinks))

	// Should have some date links
	if len(dateLinks) == 0 {
		t.Error("expected some date links in testdata")
	}

	// Log first few links
	for i, link := range links {
		if i < 5 {
			t.Logf("  Link %d: [%s](%s)", i, link.Text, link.Destination)
		}
	}
}

func TestExtractLinksFromMultipleSections(t *testing.T) {
	content := `# Section 1

Link in section 1: [yesterday](2025-01-05)

## Subsection 1.1

Another [link](2025-01-06) here.

# Section 2

And [one more](2025-01-07) in section 2.
`

	p := NewParser()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	links := doc.ExtractLinks()

	if len(links) != 3 {
		t.Errorf("expected 3 links, got %d", len(links))
	}

	// Verify all are date links
	for _, link := range links {
		if !link.IsDateLink() {
			t.Errorf("expected %q to be a date link", link.Destination)
		}
	}
}

func TestLinkLineNumbers(t *testing.T) {
	content := `# First line

Second line has a [link](dest1).

Fourth line.

Sixth line has [another](dest2).
`

	p := NewParser()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	links := doc.ExtractLinks()

	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(links))
	}

	// Line numbers should be set
	for i, link := range links {
		if link.Line == 0 {
			t.Errorf("link %d has no line number", i)
		}
		t.Logf("Link %d at line %d: [%s](%s)", i, link.Line, link.Text, link.Destination)
	}
}
