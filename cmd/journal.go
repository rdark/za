package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rdark/za/internal/markdown"
	"github.com/rdark/za/internal/notes"
	"github.com/spf13/cobra"
)

var journalWorkDoneCmd = &cobra.Command{
	Use:   "journal-work-done [date]",
	Short: "Extract work completed from journal entries",
	Long: `Extract work completed sections from a journal entry for the specified date.

If no date is provided, uses today's date.
Date format: YYYY-MM-DD

If the exact date is not found, searches backwards within the configured
search window (default: 30 days) to find the most recent entry.

The command extracts sections matching the configured work_done_sections
(default: "Work Completed", "Worked On").`,
	Args: cobra.MaximumNArgs(1),
	RunE: runJournalWorkDone,
}

func init() {
	rootCmd.AddCommand(journalWorkDoneCmd)
}

func runJournalWorkDone(cmd *cobra.Command, args []string) error {
	// Parse date argument
	var targetDate time.Time
	var err error

	if len(args) > 0 {
		targetDate, err = time.Parse(notes.DateFormat, args[0])
		if err != nil {
			return fmt.Errorf("invalid date format (expected YYYY-MM-DD): %w", err)
		}
	} else {
		targetDate = time.Now()
	}

	// Get journal directory
	journalDir, err := cfg.JournalDir()
	if err != nil {
		return fmt.Errorf("failed to get journal directory: %w", err)
	}

	// Find journal file
	journalPath, err := notes.FindNoteByDate(
		targetDate,
		notes.NoteTypeJournal,
		journalDir,
		cfg.SearchWindowDays,
	)
	if err != nil {
		return fmt.Errorf("failed to find journal entry: %w", err)
	}

	// Parse journal file
	parser := markdown.NewParser()
	doc, err := parser.ParseFile(journalPath)
	if err != nil {
		return fmt.Errorf("failed to parse journal: %w", err)
	}

	// Extract work done sections
	sections := doc.FindSectionsByHeadings(cfg.Journal.WorkDoneSections)

	if len(sections) == 0 {
		fmt.Fprintf(os.Stderr, "No work done sections found in %s\n", journalPath)
		fmt.Fprintf(os.Stderr, "Looking for sections: %v\n", cfg.Journal.WorkDoneSections)
		return nil
	}

	// Output the extracted sections
	for _, section := range sections {
		fmt.Printf("# %s\n\n", section.Heading.Text)
		fmt.Print(strings.TrimSpace(section.Content))
		fmt.Printf("\n\n")
	}

	return nil
}
