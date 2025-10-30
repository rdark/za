package markdown

import (
	"path/filepath"
	"testing"

	"github.com/yuin/goldmark/ast"
)

func TestNewParser(t *testing.T) {
	p := NewParser()
	if p == nil {
		t.Fatal("NewParser() returned nil")
	}
	if p.md == nil {
		t.Fatal("Parser markdown instance is nil")
	}
}

func TestParseBasicMarkdown(t *testing.T) {
	content := `---
title: test-document
date: 2025-01-01
tags: ["test", "example"]
---

# Main Heading

This is a paragraph with some text.

## Subheading

* List item 1
* List item 2
`

	p := NewParser()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	// Check basic document properties
	if doc.FilePath != "test.md" {
		t.Errorf("expected FilePath 'test.md', got %s", doc.FilePath)
	}

	if doc.AST == nil {
		t.Fatal("AST is nil")
	}

	if len(doc.Metadata) == 0 {
		t.Fatal("Metadata is empty")
	}
}

func TestParseFrontmatter(t *testing.T) {
	content := `---
title: my-note
date: January 1, 2025
tags: ["daily", "journal"]
number: 42
---

# Content
`

	p := NewParser()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	// Test string metadata
	title, ok := doc.GetMetadataString("title")
	if !ok {
		t.Error("expected to find 'title' metadata")
	}
	if title != "my-note" {
		t.Errorf("expected title 'my-note', got %s", title)
	}

	// Test string slice metadata
	tags, ok := doc.GetMetadataStringSlice("tags")
	if !ok {
		t.Error("expected to find 'tags' metadata")
	}
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
	if tags[0] != "daily" || tags[1] != "journal" {
		t.Errorf("unexpected tags: %v", tags)
	}
}

func TestParseWithoutFrontmatter(t *testing.T) {
	content := `# Simple Document

Just some content without frontmatter.
`

	p := NewParser()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	// Should have empty metadata map
	if doc.Metadata == nil {
		t.Error("Metadata should not be nil")
	}

	// Should still parse the document
	if doc.AST == nil {
		t.Error("AST should not be nil")
	}
}

func TestGetHeadings(t *testing.T) {
	content := `---
title: test
---

# Level 1 Heading

Some text here.

## Level 2 Heading

More text.

### Level 3 Heading

Even more text.

## Another Level 2

Final text.
`

	p := NewParser()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	headings := doc.GetHeadings()

	expectedHeadings := []struct {
		level int
		text  string
	}{
		{1, "Level 1 Heading"},
		{2, "Level 2 Heading"},
		{3, "Level 3 Heading"},
		{2, "Another Level 2"},
	}

	if len(headings) != len(expectedHeadings) {
		t.Fatalf("expected %d headings, got %d", len(expectedHeadings), len(headings))
	}

	for i, expected := range expectedHeadings {
		if headings[i].Level != expected.level {
			t.Errorf("heading %d: expected level %d, got %d", i, expected.level, headings[i].Level)
		}
		if headings[i].Text != expected.text {
			t.Errorf("heading %d: expected text %q, got %q", i, expected.text, headings[i].Text)
		}
	}
}

func TestWalkAST(t *testing.T) {
	content := `# Heading

Paragraph text.

* List item
`

	p := NewParser()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	// Count different node types
	nodeCount := 0
	headingCount := 0
	paragraphCount := 0
	listCount := 0

	doc.WalkAST(func(node ast.Node, entering bool) ast.WalkStatus {
		if entering {
			nodeCount++
			switch node.Kind() {
			case ast.KindHeading:
				headingCount++
			case ast.KindParagraph:
				paragraphCount++
			case ast.KindList:
				listCount++
			}
		}
		return ast.WalkContinue
	})

	if headingCount != 1 {
		t.Errorf("expected 1 heading, got %d", headingCount)
	}
	if paragraphCount != 1 {
		t.Errorf("expected 1 paragraph, got %d", paragraphCount)
	}
	if listCount != 1 {
		t.Errorf("expected 1 list, got %d", listCount)
	}
	if nodeCount == 0 {
		t.Error("expected to walk some nodes")
	}
}

func TestParseRealTestData(t *testing.T) {
	// Test with actual testdata file
	testFile := "../../testdata/journal/2025-01-06.md"

	p := NewParser()
	doc, err := p.ParseFile(testFile)
	if err != nil {
		t.Skipf("Skipping test, testdata file not accessible: %v", err)
		return
	}

	// Verify frontmatter exists
	if len(doc.Metadata) == 0 {
		t.Error("expected metadata from testdata file")
	}

	// Verify we can get title
	title, ok := doc.GetMetadataString("title")
	if !ok {
		t.Error("expected to find title in metadata")
	}
	if title == "" {
		t.Error("title should not be empty")
	}

	// Verify we can get headings
	headings := doc.GetHeadings()
	if len(headings) == 0 {
		t.Error("expected to find headings in testdata file")
	}

	t.Logf("Parsed testdata file successfully:")
	t.Logf("  Title: %s", title)
	t.Logf("  Headings: %d", len(headings))
}

func TestGetNodeText(t *testing.T) {
	content := `# This is a heading with **bold** and *italic* text`

	p := NewParser()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	headings := doc.GetHeadings()
	if len(headings) == 0 {
		t.Fatal("expected at least one heading")
	}

	// Get text from the heading (should include formatted text)
	text := doc.GetNodeText(headings[0].Node)

	// Should contain all the text, without markdown formatting in node text
	if text == "" {
		t.Error("expected non-empty text from heading")
	}
}

func TestParseFileNotFound(t *testing.T) {
	p := NewParser()
	_, err := p.ParseFile("/nonexistent/file.md")
	if err == nil {
		t.Error("ParseFile() should fail for non-existent file")
	}
}

func TestParseMultipleDocuments(t *testing.T) {
	// Test parsing multiple documents (simulating concurrent parsing)
	p := NewParser()

	content1 := `# Document 1`
	content2 := `# Document 2`

	doc1, err := p.Parse("doc1.md", []byte(content1))
	if err != nil {
		t.Fatalf("Parse doc1 failed: %v", err)
	}

	doc2, err := p.Parse("doc2.md", []byte(content2))
	if err != nil {
		t.Fatalf("Parse doc2 failed: %v", err)
	}

	// Verify documents are independent
	if doc1.FilePath == doc2.FilePath {
		t.Error("documents should have different file paths")
	}

	headings1 := doc1.GetHeadings()
	headings2 := doc2.GetHeadings()

	if headings1[0].Text == headings2[0].Text {
		// This is actually expected, but check they're different objects
		if headings1[0].Node == headings2[0].Node {
			t.Error("headings should be different node instances")
		}
	}
}

func TestGetMetadataTypes(t *testing.T) {
	content := `---
string_field: "hello"
number_field: 42
bool_field: true
list_field: ["one", "two", "three"]
---

# Content
`

	p := NewParser()
	doc, err := p.Parse("test.md", []byte(content))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	// Test various metadata types
	strVal, ok := doc.GetMetadataString("string_field")
	if !ok || strVal != "hello" {
		t.Errorf("expected string_field='hello', got '%s', ok=%v", strVal, ok)
	}

	// Test list
	listVal, ok := doc.GetMetadataStringSlice("list_field")
	if !ok || len(listVal) != 3 {
		t.Errorf("expected list_field with 3 items, got %v, ok=%v", listVal, ok)
	}

	// Test non-existent key
	_, ok = doc.GetMetadataString("nonexistent")
	if ok {
		t.Error("should not find nonexistent key")
	}
}

// TestParseAllTestData tests parsing all testdata files
func TestParseAllTestData(t *testing.T) {
	testFiles := []string{
		"../../testdata/journal/2025-01-06.md",
		"../../testdata/journal/2025-01-07.md",
		"../../testdata/standup/2025-01-06.md",
		"../../testdata/standup/2025-01-07.md",
	}

	p := NewParser()

	for _, file := range testFiles {
		t.Run(filepath.Base(file), func(t *testing.T) {
			doc, err := p.ParseFile(file)
			if err != nil {
				t.Skipf("Skipping, file not accessible: %v", err)
				return
			}

			// Basic validation
			if doc.AST == nil {
				t.Error("AST should not be nil")
			}

			// Should have some headings
			headings := doc.GetHeadings()
			if len(headings) == 0 {
				t.Error("expected at least one heading")
			}

			t.Logf("Successfully parsed %s: %d headings", filepath.Base(file), len(headings))
		})
	}
}
