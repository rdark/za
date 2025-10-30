package markdown

import (
	"strings"
)

// Section represents a section of a document with a heading and its content
type Section struct {
	// Heading is the section heading
	Heading Heading

	// Content is the raw content of the section (everything between this heading and the next)
	Content string
}

// ExtractSections extracts all sections from a document
// A section is defined as a heading and all content until the next heading
func (doc *Document) ExtractSections() []Section {
	// Use simple line-based extraction to avoid duplication
	return doc.ExtractSectionsSimple()
}

// FindSectionByHeading finds a section by its heading text (case-insensitive)
func (doc *Document) FindSectionByHeading(headingText string) *Section {
	sections := doc.ExtractSections()
	normalizedSearch := strings.ToLower(strings.TrimSpace(headingText))

	for _, section := range sections {
		normalizedHeading := strings.ToLower(strings.TrimSpace(section.Heading.Text))
		if normalizedHeading == normalizedSearch {
			return &section
		}
	}

	return nil
}

// FindSectionsByHeadings finds multiple sections by their heading texts (case-insensitive)
// Returns sections in the order they appear in the document
func (doc *Document) FindSectionsByHeadings(headingTexts []string) []Section {
	if len(headingTexts) == 0 {
		return []Section{}
	}

	// Normalize search terms
	searchTerms := make(map[string]bool)
	for _, text := range headingTexts {
		searchTerms[strings.ToLower(strings.TrimSpace(text))] = true
	}

	// Find matching sections
	var matchingSections []Section
	sections := doc.ExtractSections()

	for _, section := range sections {
		normalizedHeading := strings.ToLower(strings.TrimSpace(section.Heading.Text))
		if searchTerms[normalizedHeading] {
			matchingSections = append(matchingSections, section)
		}
	}

	return matchingSections
}
