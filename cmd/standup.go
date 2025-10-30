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

var standupWorkDoneCmd = &cobra.Command{
	Use:   "standup-work-done [date]",
	Short: "Extract work done from standup entries",
	Long: `Extract work done section from a standup entry for the specified date.

If no date is provided, uses today's date.
Date format: YYYY-MM-DD

If the exact date is not found, searches backwards within the configured
search window (default: 30 days) to find the most recent entry.

The command extracts the section matching the configured work_done_section
(default: "Worked on yesterday").`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStandupWorkDone,
}

func init() {
	rootCmd.AddCommand(standupWorkDoneCmd)
}

func runStandupWorkDone(cmd *cobra.Command, args []string) error {
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

	// Get standup directory
	standupDir, err := cfg.StandupDir()
	if err != nil {
		return fmt.Errorf("failed to get standup directory: %w", err)
	}

	// Find standup file
	standupPath, err := notes.FindNoteByDate(
		targetDate,
		notes.NoteTypeStandup,
		standupDir,
		cfg.SearchWindowDays,
	)
	if err != nil {
		return fmt.Errorf("failed to find standup entry: %w", err)
	}

	// Parse standup file
	parser := markdown.NewParser()
	doc, err := parser.ParseFile(standupPath)
	if err != nil {
		return fmt.Errorf("failed to parse standup: %w", err)
	}

	// Extract work done section
	section := doc.FindSectionByHeading(cfg.Standup.WorkDoneSection)

	if section == nil {
		fmt.Fprintf(os.Stderr, "No work done section found in %s\n", standupPath)
		fmt.Fprintf(os.Stderr, "Looking for section: %q\n", cfg.Standup.WorkDoneSection)
		return nil
	}

	// Output the extracted section
	fmt.Printf("# %s\n\n", section.Heading.Text)
	fmt.Print(strings.TrimSpace(section.Content))
	fmt.Printf("\n\n")

	return nil
}
