package markdown

import (
	"regexp"
	"strings"

	"github.com/yuin/goldmark/ast"
)

// Link represents a markdown link found in a document
type Link struct {
	// Text is the link text (what appears between [])
	Text string

	// Destination is the link target (what appears between ())
	Destination string

	// Line is the line number where the link appears (1-indexed)
	Line int

	// Node is the AST node for this link
	Node *ast.Link
}

// ExtractLinks extracts all markdown links from the document
func (doc *Document) ExtractLinks() []Link {
	var links []Link

	doc.WalkAST(func(node ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}

		if linkNode, ok := node.(*ast.Link); ok {
			// Get link text
			text := doc.GetNodeText(linkNode)

			// Get destination
			destination := string(linkNode.Destination)

			// Get line number by finding parent block node
			line := 0
			parent := linkNode.Parent()
			for parent != nil {
				if parent.Lines().Len() > 0 {
					lineSegment := parent.Lines().At(0)
					// Count newlines up to this point to get approximate line number
					line = countLines(doc.Source[:lineSegment.Start]) + 1
					break
				}
				parent = parent.Parent()
			}

			links = append(links, Link{
				Text:        text,
				Destination: destination,
				Line:        line,
				Node:        linkNode,
			})
		}

		return ast.WalkContinue
	})

	return links
}

// countLines counts the number of newlines in a byte slice
func countLines(data []byte) int {
	count := 0
	for _, b := range data {
		if b == '\n' {
			count++
		}
	}
	return count
}

// IsDateLink returns true if the link destination looks like a date (YYYY-MM-DD)
func (l *Link) IsDateLink() bool {
	// Match YYYY-MM-DD pattern
	matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}(\.md)?$`, l.Destination)
	if matched {
		return true
	}

	// Also check for relative paths like ../journal/YYYY-MM-DD.md
	matched, _ = regexp.MatchString(`\.\./[^/]+/\d{4}-\d{2}-\d{2}(\.md)?$`, l.Destination)
	return matched
}

// IsRelativeLink returns true if the link is a relative path
func (l *Link) IsRelativeLink() bool {
	return strings.HasPrefix(l.Destination, ".") ||
		!strings.Contains(l.Destination, "://")
}

// IsExternalLink returns true if the link is an external URL
func (l *Link) IsExternalLink() bool {
	return strings.Contains(l.Destination, "://")
}

// GetDateFromDestination extracts the date portion from a link destination
// Returns the date string (YYYY-MM-DD) or empty string if not a date link
func (l *Link) GetDateFromDestination() string {
	datePattern := regexp.MustCompile(`(\d{4}-\d{2}-\d{2})`)
	matches := datePattern.FindStringSubmatch(l.Destination)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// GetNoteTypeFromDestination tries to determine the note type from the link destination
// Returns "journal", "standup", or "" if unknown
func (l *Link) GetNoteTypeFromDestination() string {
	dest := strings.ToLower(l.Destination)

	if strings.Contains(dest, "/journal/") || strings.HasPrefix(dest, "journal/") {
		return "journal"
	}
	if strings.Contains(dest, "/standup/") || strings.HasPrefix(dest, "standup/") {
		return "standup"
	}

	return ""
}

// FilterLinks filters links based on a predicate function
func FilterLinks(links []Link, predicate func(Link) bool) []Link {
	var filtered []Link
	for _, link := range links {
		if predicate(link) {
			filtered = append(filtered, link)
		}
	}
	return filtered
}
