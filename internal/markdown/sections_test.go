package markdown

import (
	"strings"
	"testing"
)

func TestExtractSections(t *testing.T) {
	content := `---
title: test
---

# Section 1

This is content for section 1.
It has multiple lines.

## Section 2

Content for section 2.

* List item 1
* List item 2

### Section 3

Final section content.
`

	p := NewParser()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	sections := doc.ExtractSections()

	// Should have 3 sections
	if len(sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(sections))
	}

	// Check section headings
	expectedHeadings := []string{"Section 1", "Section 2", "Section 3"}
	for i, expected := range expectedHeadings {
		if sections[i].Heading.Text != expected {
			t.Errorf("section %d: expected heading %q, got %q", i, expected, sections[i].Heading.Text)
		}
	}

	// Check section levels
	expectedLevels := []int{1, 2, 3}
	for i, expected := range expectedLevels {
		if sections[i].Heading.Level != expected {
			t.Errorf("section %d: expected level %d, got %d", i, expected, sections[i].Heading.Level)
		}
	}

	// Check that sections have content
	for i, section := range sections {
		if strings.TrimSpace(section.Content) == "" {
			t.Errorf("section %d (%s) has empty content", i, section.Heading.Text)
		}
		t.Logf("Section %d content: %q", i, strings.TrimSpace(section.Content))
	}
}

func TestExtractSectionsWithLists(t *testing.T) {
	content := `# Work Completed

* Task 1
  * Subtask 1a
  * Subtask 1b
* Task 2
* Task 3

# Next Section

Some text.
`

	p := NewParser()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	sections := doc.ExtractSections()
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}

	// Check that list is rendered
	firstSection := sections[0]
	if !strings.Contains(firstSection.Content, "* Task 1") {
		t.Error("expected list content in first section")
	}
	if !strings.Contains(firstSection.Content, "  * Subtask 1a") {
		t.Error("expected nested list content in first section")
	}

	t.Logf("First section content:\n%s", firstSection.Content)
}

func TestExtractSectionsWithLinks(t *testing.T) {
	content := `# Links

* [Yesterday](2025-01-05)
* [Tomorrow](2025-01-07)
* [External](https://example.com)

# Other

Text here.
`

	p := NewParser()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	sections := doc.ExtractSections()
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}

	// Check that links are preserved
	firstSection := sections[0]
	if !strings.Contains(firstSection.Content, "[Yesterday](2025-01-05)") {
		t.Error("expected link to be preserved in content")
	}
	if !strings.Contains(firstSection.Content, "[External](https://example.com)") {
		t.Error("expected external link to be preserved")
	}

	t.Logf("Links section content:\n%s", firstSection.Content)
}

func TestExtractSectionsWithCodeBlocks(t *testing.T) {
	content := "# Code Example\n\n```go\nfunc main() {\n    fmt.Println(\"hello\")\n}\n```\n\n# Next\n\nText.\n"

	p := NewParser()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	sections := doc.ExtractSections()
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}

	// Check that code block is preserved
	firstSection := sections[0]
	if !strings.Contains(firstSection.Content, "```") {
		t.Error("expected code block markers in content")
	}
	if !strings.Contains(firstSection.Content, "func main()") {
		t.Error("expected code content in section")
	}

	t.Logf("Code section content:\n%s", firstSection.Content)
}

func TestExtractSectionsWithBlockquotes(t *testing.T) {
	content := `# Notes

> This is a quoted text
> spanning multiple lines

Regular text after quote.

# Next

More text.
`

	p := NewParser()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	sections := doc.ExtractSections()
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}

	// Check that blockquote is rendered
	firstSection := sections[0]
	if !strings.Contains(firstSection.Content, ">") {
		t.Error("expected blockquote marker in content")
	}

	t.Logf("Blockquote section content:\n%s", firstSection.Content)
}

func TestFindSectionByHeading(t *testing.T) {
	content := `# First Heading

Content 1

# Second Heading

Content 2

# Third Heading

Content 3
`

	p := NewParser()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	tests := []struct {
		name       string
		search     string
		shouldFind bool
		expectText string
	}{
		{
			name:       "exact match",
			search:     "Second Heading",
			shouldFind: true,
			expectText: "Second Heading",
		},
		{
			name:       "case insensitive",
			search:     "second heading",
			shouldFind: true,
			expectText: "Second Heading",
		},
		{
			name:       "with extra spaces",
			search:     "  Second Heading  ",
			shouldFind: true,
			expectText: "Second Heading",
		},
		{
			name:       "not found",
			search:     "Nonexistent Heading",
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			section := doc.FindSectionByHeading(tt.search)

			if tt.shouldFind {
				if section == nil {
					t.Errorf("FindSectionByHeading(%q) returned nil, expected section", tt.search)
					return
				}
				if section.Heading.Text != tt.expectText {
					t.Errorf("expected heading %q, got %q", tt.expectText, section.Heading.Text)
				}
				if strings.TrimSpace(section.Content) == "" {
					t.Error("section content is empty")
				}
			} else {
				if section != nil {
					t.Errorf("FindSectionByHeading(%q) found section, expected nil", tt.search)
				}
			}
		})
	}
}

func TestFindSectionsByHeadings(t *testing.T) {
	content := `# Work Completed

Task 1 done
Task 2 done

# Worked On

Task 3 in progress

# Meetings

Met with team

# Thoughts

Random thoughts
`

	p := NewParser()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	// Find multiple sections
	sections := doc.FindSectionsByHeadings([]string{"Work Completed", "Worked On"})

	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}

	// Check headings
	if sections[0].Heading.Text != "Work Completed" {
		t.Errorf("expected first section 'Work Completed', got %q", sections[0].Heading.Text)
	}
	if sections[1].Heading.Text != "Worked On" {
		t.Errorf("expected second section 'Worked On', got %q", sections[1].Heading.Text)
	}

	// Check content exists
	for i, section := range sections {
		if strings.TrimSpace(section.Content) == "" {
			t.Errorf("section %d has empty content", i)
		}
	}
}

func TestFindSectionsByHeadingsEmpty(t *testing.T) {
	content := `# Heading

Content
`

	p := NewParser()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	// Empty search
	sections := doc.FindSectionsByHeadings([]string{})
	if len(sections) != 0 {
		t.Error("expected empty result for empty search")
	}

	// No matches
	sections = doc.FindSectionsByHeadings([]string{"Nonexistent"})
	if len(sections) != 0 {
		t.Error("expected empty result for no matches")
	}
}

func TestExtractSectionsRealTestData(t *testing.T) {
	testFile := "../../testdata/journal/2025-01-06.md"

	p := NewParser()
	doc, err := p.ParseFile(testFile)
	if err != nil {
		t.Skipf("Skipping, testdata file not accessible: %v", err)
		return
	}

	sections := doc.ExtractSections()

	if len(sections) == 0 {
		t.Fatal("expected sections from testdata file")
	}

	t.Logf("Found %d sections in testdata file", len(sections))

	// Log first few section headings
	for i, section := range sections {
		if i < 5 {
			t.Logf("  Section %d: %q (level %d)", i, section.Heading.Text, section.Heading.Level)
		}
	}

	// Try to find known sections
	workCompleted := doc.FindSectionByHeading("Work Completed")
	if workCompleted == nil {
		t.Error("expected to find 'Work Completed' section")
	} else {
		t.Logf("Work Completed section content length: %d", len(workCompleted.Content))
		if !strings.Contains(workCompleted.Content, "*") {
			t.Error("expected list items in Work Completed section")
		}
	}
}

func TestFindMultipleSectionsRealTestData(t *testing.T) {
	testFile := "../../testdata/journal/2025-01-06.md"

	p := NewParser()
	doc, err := p.ParseFile(testFile)
	if err != nil {
		t.Skipf("Skipping, testdata file not accessible: %v", err)
		return
	}

	// Find work-related sections (matching config defaults)
	sections := doc.FindSectionsByHeadings([]string{"Work Completed", "Worked On"})

	if len(sections) == 0 {
		t.Error("expected to find work sections")
	}

	t.Logf("Found %d work sections", len(sections))
	for i, section := range sections {
		t.Logf("  %d: %s (%d chars)", i, section.Heading.Text, len(section.Content))
	}
}

func TestExtractSectionsEmptyDocument(t *testing.T) {
	content := `---
title: empty
---

No headings here, just text.
`

	p := NewParser()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	sections := doc.ExtractSections()

	// Should return empty slice for document with no headings
	if len(sections) != 0 {
		t.Errorf("expected 0 sections for document without headings, got %d", len(sections))
	}
}

func TestExtractSectionsPreservesFormatting(t *testing.T) {
	content := `# Formatting Test

This has **bold** and *italic* and ` + "`code`" + `.

Also a [link](https://example.com).

* List with **bold item**
* And *italic item*

# Next

Text.
`

	p := NewParser()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	sections := doc.ExtractSections()
	if len(sections) == 0 {
		t.Fatal("expected at least one section")
	}

	firstSection := sections[0]

	// Check formatting is preserved
	checks := []string{
		"**bold**",
		"*italic*",
		"`code`",
		"[link](https://example.com)",
	}

	for _, check := range checks {
		if !strings.Contains(firstSection.Content, check) {
			t.Errorf("expected %q in content", check)
		}
	}

	t.Logf("Formatted section content:\n%s", firstSection.Content)
}
