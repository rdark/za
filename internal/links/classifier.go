// Package links provides functionality for classifying, resolving, and fixing
// links in markdown documents. It handles temporal links (Yesterday/Tomorrow),
// cross-reference links (Journal/Standup), and can resolve stale date links
// to point to actual existing files.
package links

import (
	"strings"

	"github.com/rdark/za/internal/config"
	"github.com/rdark/za/internal/markdown"
)

// LinkType represents the type/purpose of a link
type LinkType string

const (
	// LinkTypeTemporalPrevious represents links to previous entries (Yesterday, Previous, etc.)
	LinkTypeTemporalPrevious LinkType = "temporal_previous"

	// LinkTypeTemporalNext represents links to next entries (Tomorrow, Next, etc.)
	LinkTypeTemporalNext LinkType = "temporal_next"

	// LinkTypeCrossReference represents links between different note types (Journal <-> Standup)
	LinkTypeCrossReference LinkType = "cross_reference"

	// LinkTypeExternal represents external URLs
	LinkTypeExternal LinkType = "external"

	// LinkTypeOther represents other types of links (wiki links, etc.)
	LinkTypeOther LinkType = "other"
)

// ClassifiedLink represents a link with its classification
type ClassifiedLink struct {
	// Link is the original markdown link
	Link markdown.Link

	// Type is the classified type of the link
	Type LinkType

	// TargetNoteType is the type of note this link points to (if applicable)
	TargetNoteType string
}

// Classifier classifies markdown links
type Classifier struct {
	cfg *config.Config
}

// NewClassifier creates a new link classifier
func NewClassifier(cfg *config.Config) *Classifier {
	return &Classifier{cfg: cfg}
}

// Classify classifies a single link
func (c *Classifier) Classify(link markdown.Link) ClassifiedLink {
	classified := ClassifiedLink{
		Link: link,
		Type: LinkTypeOther,
	}

	// Check if it's an external link
	if link.IsExternalLink() {
		classified.Type = LinkTypeExternal
		return classified
	}

	// Check if it's a date link
	if !link.IsDateLink() {
		// Not a date link, might be wiki link or other
		return classified
	}

	// It's a date link - determine if it's temporal or cross-reference
	linkText := strings.ToLower(strings.TrimSpace(link.Text))

	// Check for temporal previous synonyms
	if c.matchesAny(linkText, c.cfg.Journal.LinkPreviousTitles) ||
		c.matchesAny(linkText, c.cfg.Standup.LinkPreviousTitles) {
		classified.Type = LinkTypeTemporalPrevious
		// Try to determine target note type from destination
		classified.TargetNoteType = link.GetNoteTypeFromDestination()
		return classified
	}

	// Check for temporal next synonyms
	if c.matchesAny(linkText, c.cfg.Journal.LinkNextTitles) ||
		c.matchesAny(linkText, c.cfg.Standup.LinkNextTitles) {
		classified.Type = LinkTypeTemporalNext
		classified.TargetNoteType = link.GetNoteTypeFromDestination()
		return classified
	}

	// Check for cross-reference patterns
	if c.isCrossReference(linkText) {
		classified.Type = LinkTypeCrossReference
		classified.TargetNoteType = link.GetNoteTypeFromDestination()
		return classified
	}

	return classified
}

// ClassifyAll classifies all links in a list
func (c *Classifier) ClassifyAll(links []markdown.Link) []ClassifiedLink {
	classified := make([]ClassifiedLink, 0, len(links))
	for _, link := range links {
		classified = append(classified, c.Classify(link))
	}
	return classified
}

// matchesAny checks if the text matches any of the provided patterns (case-insensitive)
func (c *Classifier) matchesAny(text string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.ToLower(strings.TrimSpace(pattern)) == text {
			return true
		}
	}
	return false
}

// isCrossReference checks if the link text indicates a cross-reference
func (c *Classifier) isCrossReference(linkText string) bool {
	// Common cross-reference patterns
	crossRefPatterns := []string{
		"standup",
		"journal",
		"daily",
		"daily log",
	}

	for _, pattern := range crossRefPatterns {
		if strings.Contains(linkText, pattern) {
			return true
		}
	}

	return false
}

// FilterByType filters classified links by type
func FilterByType(links []ClassifiedLink, linkType LinkType) []ClassifiedLink {
	var filtered []ClassifiedLink
	for _, link := range links {
		if link.Type == linkType {
			filtered = append(filtered, link)
		}
	}
	return filtered
}

// NeedsFixing returns true if a classified link might need fixing
// Temporal and cross-reference links with date destinations are candidates for fixing
func (l *ClassifiedLink) NeedsFixing() bool {
	switch l.Type {
	case LinkTypeTemporalPrevious, LinkTypeTemporalNext, LinkTypeCrossReference:
		// These types might need fixing if they have a date
		return l.Link.IsDateLink()
	default:
		return false
	}
}

// IsNextLink returns true if this is a temporal "next" link
func (l *ClassifiedLink) IsNextLink() bool {
	return l.Type == LinkTypeTemporalNext
}
