package markdown

import (
	"bufio"
	"bytes"
	"strings"
)

// ExtractSectionsSimple extracts sections using a simpler line-based approach
func (doc *Document) ExtractSectionsSimple() []Section {
	var sections []Section

	// Get all headings
	headings := doc.GetHeadings()
	if len(headings) == 0 {
		return sections
	}

	// For each heading, extract content until next heading
	for i, heading := range headings {
		// Get start line of content (after heading)
		startLine := heading.Node.Lines().At(0).Start

		// Get end line (start of next heading or end of document)
		var endLine int
		if i < len(headings)-1 {
			endLine = headings[i+1].Node.Lines().At(0).Start
		} else {
			endLine = len(doc.Source)
		}

		// Extract content between these lines
		content := extractContentBetween(doc.Source, startLine, endLine)

		sections = append(sections, Section{
			Heading: heading,
			Content: content,
		})
	}

	return sections
}

// extractContentBetween extracts content from source between start and end byte positions
func extractContentBetween(source []byte, start, end int) string {
	if start >= len(source) {
		return ""
	}
	if end > len(source) {
		end = len(source)
	}

	// Find the start of the next line after the heading
	scanner := bufio.NewScanner(bytes.NewReader(source[start:end]))

	var lines []string
	firstLine := true
	for scanner.Scan() {
		line := scanner.Text()

		// Skip the first line (the heading itself)
		if firstLine {
			firstLine = false
			continue
		}

		// Stop if we hit another heading (line starting with #)
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "#") {
			break
		}

		lines = append(lines, line)
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}
