package markdown

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddTagToFile(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		tag            string
		expectAdded    bool
		expectInResult string
	}{
		{
			name: "add tag to existing tags array",
			content: `---
title: Test Document
tags: ["daily", "journal"]
---

# Content
Some content here`,
			tag:            "company:acme",
			expectAdded:    true,
			expectInResult: `company:acme`, // Just check the tag is present
		},
		{
			name: "tag already exists",
			content: `---
title: Test Document
tags: ["daily", "journal", "company:acme"]
---

# Content`,
			tag:            "company:acme",
			expectAdded:    false,
			expectInResult: `company:acme`, // Check tag is still present
		},
		{
			name: "no frontmatter",
			content: `# Content
No frontmatter here`,
			tag:         "company:acme",
			expectAdded: false,
		},
		{
			name: "no tags field",
			content: `---
title: Test Document
date: 2025-01-01
---

# Content`,
			tag:         "company:acme",
			expectAdded: false,
		},
		{
			name: "empty tags array",
			content: `---
title: Test Document
tags: []
---

# Content`,
			tag:            "company:acme",
			expectAdded:    true,
			expectInResult: "company:acme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(filePath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Add tag
			added, err := AddTagToFile(filePath, tt.tag)
			if err != nil {
				t.Fatalf("AddTagToFile failed: %v", err)
			}

			if added != tt.expectAdded {
				t.Errorf("Expected added=%v, got %v", tt.expectAdded, added)
			}

			// Read result
			result, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read result file: %v", err)
			}

			resultStr := string(result)

			if tt.expectInResult != "" {
				if !strings.Contains(resultStr, tt.expectInResult) {
					t.Errorf("Expected result to contain %q\nGot:\n%s", tt.expectInResult, resultStr)
				}
			}

			// Verify content is preserved
			if tt.expectAdded && strings.Contains(tt.content, "# Content") {
				if !strings.Contains(resultStr, "# Content") {
					t.Errorf("Content section was not preserved")
				}
			}
		})
	}
}

func TestAddTagToFile_InlineArrayFormat(t *testing.T) {
	// Test that tags are rendered in inline array format
	content := `---
title: Test Document
tags: ["daily", "journal"]
---

# Content`

	// Create temp file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Add tag
	added, err := AddTagToFile(filePath, "company:acme")
	if err != nil {
		t.Fatalf("AddTagToFile failed: %v", err)
	}

	if !added {
		t.Errorf("Expected tag to be added")
	}

	// Read result
	result, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read result file: %v", err)
	}

	resultStr := string(result)
	t.Logf("Result:\n%s", resultStr)

	// Check for inline array format - should NOT contain multi-line format
	if strings.Contains(resultStr, "tags:\n") {
		t.Errorf("Tags should be in inline format, not multi-line. Got:\n%s", resultStr)
	}

	// Should contain inline array format
	if !strings.Contains(resultStr, `tags: [`) {
		t.Errorf("Expected inline array format 'tags: ['. Got:\n%s", resultStr)
	}

	// Verify all tags are present with double quotes
	if !strings.Contains(resultStr, `"daily"`) {
		t.Errorf("Missing 'daily' tag with double quotes. Got:\n%s", resultStr)
	}
	if !strings.Contains(resultStr, `"journal"`) {
		t.Errorf("Missing 'journal' tag with double quotes. Got:\n%s", resultStr)
	}
	if !strings.Contains(resultStr, `"company:acme"`) {
		t.Errorf("Missing 'company:acme' tag with double quotes. Got:\n%s", resultStr)
	}

	// Verify consistent double-quote format (no single quotes)
	if strings.Contains(resultStr, `'daily'`) || strings.Contains(resultStr, `'journal'`) || strings.Contains(resultStr, `'company:acme'`) {
		t.Errorf("Tags should use double quotes, not single quotes. Got:\n%s", resultStr)
	}

	// Check exact format
	tagsLine := ""
	for _, line := range strings.Split(resultStr, "\n") {
		if strings.HasPrefix(line, "tags:") {
			tagsLine = line
			break
		}
	}

	if tagsLine == "" {
		t.Errorf("Could not find tags line in output")
	} else {
		t.Logf("Tags line: %s", tagsLine)
		// Check for exact expected format
		expected := `tags: ["daily", "journal", "company:acme"]`
		if tagsLine != expected {
			t.Errorf("Tags line format mismatch\nExpected: %s\nGot:      %s", expected, tagsLine)
		}
	}
}

func TestAddTagToFile_QuoteConsistency(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		tagToAdd    string
		wantContain string
		wantExact   string
	}{
		{
			name: "simple tags with double quotes",
			content: `---
title: Test
tags: ["daily", "journal"]
---
Content`,
			tagToAdd:  "company:acme",
			wantExact: `tags: ["daily", "journal", "company:acme"]`,
		},
		{
			name: "tags with special characters",
			content: `---
title: Test
tags: ["status:active", "type:task"]
---
Content`,
			tagToAdd:  "company:acme",
			wantExact: `tags: ["status:active", "type:task", "company:acme"]`,
		},
		{
			name: "single tag originally",
			content: `---
title: Test
tags: ["daily"]
---
Content`,
			tagToAdd:  "company:acme",
			wantExact: `tags: ["daily", "company:acme"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test.md")

			if err := os.WriteFile(filePath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			added, err := AddTagToFile(filePath, tt.tagToAdd)
			if err != nil {
				t.Fatalf("AddTagToFile failed: %v", err)
			}

			if !added {
				t.Errorf("Expected tag to be added")
			}

			result, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read result: %v", err)
			}

			resultStr := string(result)
			t.Logf("Full result:\n%s", resultStr)

			// Find tags line
			tagsLine := ""
			for _, line := range strings.Split(resultStr, "\n") {
				if strings.HasPrefix(line, "tags:") {
					tagsLine = line
					break
				}
			}

			if tagsLine == "" {
				t.Fatalf("Could not find tags line")
			}

			t.Logf("Tags line: %q", tagsLine)

			if tt.wantExact != "" && tagsLine != tt.wantExact {
				t.Errorf("Tags line mismatch\nWant: %q\nGot:  %q", tt.wantExact, tagsLine)
			}

			// Check for inconsistent quoting patterns
			if strings.Contains(tagsLine, "'") {
				t.Errorf("Tags contain single quotes (should be double): %s", tagsLine)
			}

			// Check for unquoted strings (except for brackets, commas, colons inside quotes)
			if strings.Contains(tagsLine, "[daily,") || strings.Contains(tagsLine, ", journal,") {
				t.Errorf("Tags appear to be unquoted or inconsistently quoted: %s", tagsLine)
			}
		})
	}
}

func TestAddTagToFile_ExactOutput(t *testing.T) {
	// Extremely detailed test to verify exact output
	content := `---
title: Test Document
date: January 1, 2025
tags: ["daily", "journal"]
---

# Content`

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Add tag
	added, err := AddTagToFile(filePath, "company:acme")
	if err != nil {
		t.Fatalf("AddTagToFile failed: %v", err)
	}

	if !added {
		t.Fatalf("Tag was not added")
	}

	// Read result
	result, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read result: %v", err)
	}

	resultStr := string(result)
	lines := strings.Split(resultStr, "\n")

	t.Log("Complete output:")
	for i, line := range lines {
		t.Logf("  Line %d: %q", i, line)
	}

	// Find and examine the tags line
	var tagsLine string
	var tagsLineNum int
	for i, line := range lines {
		if strings.HasPrefix(line, "tags:") {
			tagsLine = line
			tagsLineNum = i
			break
		}
	}

	if tagsLine == "" {
		t.Fatalf("No tags line found")
	}

	t.Logf("\nTags line (#%d): %q", tagsLineNum, tagsLine)
	t.Logf("Tags line length: %d bytes", len(tagsLine))

	// Show each character
	t.Log("Character breakdown:")
	for i, c := range tagsLine {
		t.Logf("  [%d]: %c (0x%02X)", i, c, c)
	}

	// Exact match test
	want := `tags: ["daily", "journal", "company:acme"]`
	if tagsLine != want {
		t.Errorf("\nTags line mismatch:")
		t.Errorf("  Want: %q", want)
		t.Errorf("  Got:  %q", tagsLine)
		t.Errorf("\nCharacter differences:")
		for i := 0; i < len(want) || i < len(tagsLine); i++ {
			if i >= len(want) {
				t.Errorf("  [%d]: Extra in output: %c", i, tagsLine[i])
			} else if i >= len(tagsLine) {
				t.Errorf("  [%d]: Missing from output: %c", i, want[i])
			} else if want[i] != tagsLine[i] {
				t.Errorf("  [%d]: Want %c (0x%02X), Got %c (0x%02X)", i, want[i], want[i], tagsLine[i], tagsLine[i])
			}
		}
	}
}

func TestExtractFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		expectEnd   int
	}{
		{
			name: "valid frontmatter",
			content: `---
title: Test
---
Content`,
			expectError: false,
			expectEnd:   17, // Length of frontmatter section
		},
		{
			name: "no frontmatter",
			content: `# Content
No frontmatter`,
			expectError: true,
		},
		{
			name: "unclosed frontmatter",
			content: `---
title: Test
More content without closing`,
			expectError: true,
		},
		{
			name:        "empty file",
			content:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			end, fm, err := extractFrontmatter([]byte(tt.content))

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if end == 0 {
					t.Errorf("Expected non-zero end position")
				}
				if len(fm) == 0 {
					t.Errorf("Expected non-empty frontmatter")
				}
			}
		})
	}
}
