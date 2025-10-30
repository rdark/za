package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/rdark/za/internal/links"
	"github.com/rdark/za/internal/markdown"
	"github.com/rdark/za/internal/notes"
	"github.com/spf13/cobra"
)

var (
	dryRun bool
)

var fixLinksCmd = &cobra.Command{
	Use:   "fix-links <file>",
	Short: "Fix relative date links in a note file",
	Long: `Fix relative date links in a note file by resolving them to actual entries.

This command analyzes all links in a markdown file and updates temporal links
(Yesterday, Tomorrow, etc.) and cross-reference links (Standup, Journal, etc.)
to point to actual existing notes, skipping gaps like weekends and holidays.

The command handles:
- Temporal links: Yesterday/Previous, Tomorrow/Next (with synonyms)
- Cross-references: Journal <-> Standup
- Gap handling: Skips missing days, weekends, holidays

By default, the file is modified in place. Use --dry-run to preview changes
without modifying the file.`,
	Args: cobra.ExactArgs(1),
	RunE: runFixLinks,
}

func init() {
	rootCmd.AddCommand(fixLinksCmd)
	fixLinksCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without modifying the file")
}

func runFixLinks(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Check file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Determine note type from path
	noteType, err := determineNoteType(filePath)
	if err != nil {
		return fmt.Errorf("failed to determine note type: %w", err)
	}

	// Parse date from filename
	fileDate, err := notes.ParseDateFromFilename(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse date from filename: %w", err)
	}

	// Parse the file
	parser := markdown.NewParser()
	doc, err := parser.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	// Extract all links
	allLinks := doc.ExtractLinks()

	if len(allLinks) == 0 {
		fmt.Println("No links found in file")
		return nil
	}

	// Classify links
	classifier := links.NewClassifier(cfg)
	classified := classifier.ClassifyAll(allLinks)

	// Filter to only fixable links
	fixable := make([]links.ClassifiedLink, 0)
	for _, c := range classified {
		if c.NeedsFixing() {
			fixable = append(fixable, c)
		}
	}

	if len(fixable) == 0 {
		fmt.Println("No fixable links found in file")
		return nil
	}

	fmt.Printf("Found %d fixable links\n", len(fixable))

	// Resolve links
	resolver := links.NewResolver(cfg, fileDate, noteType)
	resolved := resolver.ResolveAll(fixable)

	// Filter to links that need updating
	needsUpdate := links.FilterNeedsUpdate(resolved)

	if len(needsUpdate) == 0 {
		fmt.Println("All links are already correct!")
		return nil
	}

	fmt.Printf("\n%d links need updating:\n\n", len(needsUpdate))

	// Display changes
	for i, r := range needsUpdate {
		if r.Error != nil {
			fmt.Printf("%d. [%s](%s) - ERROR: %v\n",
				i+1,
				r.Classified.Link.Text,
				r.Classified.Link.Destination,
				r.Error,
			)
			continue
		}

		fmt.Printf("%d. [%s](%s)\n",
			i+1,
			r.Classified.Link.Text,
			r.Classified.Link.Destination,
		)
		fmt.Printf("   → %s\n",
			r.SuggestedDestination,
		)
		fmt.Printf("   Type: %s\n",
			r.Classified.Type,
		)
	}

	// If dry-run, stop here
	if dryRun {
		fmt.Println("\n[DRY RUN] No changes made")
		return nil
	}

	// Apply changes
	fmt.Println("\nApplying changes...")

	newContent, err := applyLinkFixes(doc, needsUpdate)
	if err != nil {
		return fmt.Errorf("failed to apply link fixes: %w", err)
	}

	// Write back to file
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("\n✓ Successfully updated %d links in %s\n", len(needsUpdate), filePath)

	return nil
}

// determineNoteType determines the note type from the file path by checking
// if any path component matches "journal" or "standup" (case-insensitive).
func determineNoteType(filePath string) (notes.NoteType, error) {
	// Normalize path separators and split into components
	normalizedPath := strings.ReplaceAll(filePath, "\\", "/")

	// Check each component for journal or standup
	for component := range strings.SplitSeq(normalizedPath, "/") {
		lowerComponent := strings.ToLower(component)
		switch lowerComponent {
		case "journal":
			return notes.NoteTypeJournal, nil
		case "standup":
			return notes.NoteTypeStandup, nil
		}
	}

	return "", fmt.Errorf("cannot determine note type from path: %s (expected path to contain 'journal' or 'standup' directory)", filePath)
}

// applyLinkFixes applies link fixes to the document content
func applyLinkFixes(doc *markdown.Document, fixes []links.ResolvedLink) (string, error) {
	content := string(doc.Content)

	// Sort fixes by line number (descending) to avoid position shifts
	// For now, do simple string replacement (could be improved with AST manipulation)

	for _, fix := range fixes {
		if fix.Error != nil {
			continue
		}

		// Build old and new link strings
		oldLink := fmt.Sprintf("[%s](%s)", fix.Classified.Link.Text, fix.Classified.Link.Destination)
		newLink := fmt.Sprintf("[%s](%s)", fix.Classified.Link.Text, fix.SuggestedDestination)

		// Replace (only first occurrence to be safe)
		content = strings.Replace(content, oldLink, newLink, 1)
	}

	return content, nil
}
