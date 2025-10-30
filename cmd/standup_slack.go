package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/rdark/za/internal/markdown"
	"github.com/rdark/za/internal/notes"
	"github.com/spf13/cobra"
)

var standupSlackCmd = &cobra.Command{
	Use:   "standup-slack [date]",
	Short: "Print a concise daily update in Slack-compatible markdown",
	Long: `Print a summary of yesterday's completed work and today's planned work
in a format suitable for pasting into Slack.

This command extracts:
- Completed goals and work from yesterday's journal
- Goals for today from today's journal

Examples:
  za standup-slack                    # Generate update for today
  za standup-slack 2025-01-15        # Generate update for specific date`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStandupSlack,
}

func init() {
	rootCmd.AddCommand(standupSlackCmd)
}

func runStandupSlack(cmd *cobra.Command, args []string) error {
	// Parse target date (today)
	var targetDate time.Time
	var err error
	if len(args) > 0 {
		targetDate, err = time.Parse(notes.DateFormat, args[0])
		if err != nil {
			return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
		}
	} else {
		targetDate = time.Now()
	}

	journalDir, err := cfg.JournalDir()
	if err != nil {
		return fmt.Errorf("failed to get journal directory: %w", err)
	}

	parser := markdown.NewParser()

	// Extract yesterday's work
	var yesterdayItems []string
	previousDate := targetDate.AddDate(0, 0, -1)
	prevJournalPath, err := notes.FindNoteByDate(previousDate, notes.NoteTypeJournal, journalDir, cfg.SearchWindowDays)
	if err == nil {
		prevDoc, err := parser.ParseFile(prevJournalPath)
		if err == nil {
			// Extract completed goals
			prevGoalsSection := prevDoc.FindSectionByHeading("Goals of the Day")
			if prevGoalsSection != nil && strings.TrimSpace(prevGoalsSection.Content) != "" {
				items := markdown.ParseGoalItems(prevGoalsSection.Content)
				for _, item := range items {
					if item.HasCheckbox && item.Checked {
						yesterdayItems = append(yesterdayItems, item.Text)
					}
				}
			}

			// Extract work sections
			workSections := prevDoc.FindSectionsByHeadings(cfg.Journal.WorkDoneSections)
			for _, section := range workSections {
				sectionContent := strings.TrimSpace(section.Content)
				if sectionContent != "" {
					// Parse bullet points from work section
					lines := strings.Split(sectionContent, "\n")
					for _, line := range lines {
						trimmed := strings.TrimSpace(line)
						// Extract bullet points (both * and -)
						if strings.HasPrefix(trimmed, "* ") {
							yesterdayItems = append(yesterdayItems, strings.TrimPrefix(trimmed, "* "))
						} else if strings.HasPrefix(trimmed, "- ") {
							yesterdayItems = append(yesterdayItems, strings.TrimPrefix(trimmed, "- "))
						}
					}
				}
			}
		}
	}

	// Extract today's goals
	var todayItems []string
	todayJournalPath, err := notes.FindNoteByDate(targetDate, notes.NoteTypeJournal, journalDir, cfg.SearchWindowDays)
	if err == nil {
		// Verify this is actually today's journal
		foundDate, err := notes.ParseDateFromFilename(todayJournalPath)
		if err == nil {
			// Compare just the date parts
			targetY, targetM, targetD := targetDate.Date()
			foundY, foundM, foundD := foundDate.Date()
			if targetY == foundY && targetM == foundM && targetD == foundD {
				todayDoc, err := parser.ParseFile(todayJournalPath)
				if err == nil {
					todayGoalsSection := todayDoc.FindSectionByHeading("Goals of the Day")
					if todayGoalsSection != nil && strings.TrimSpace(todayGoalsSection.Content) != "" {
						items := markdown.ParseGoalItems(todayGoalsSection.Content)
						for _, item := range items {
							if item.Text != "" {
								todayItems = append(todayItems, item.Text)
							}
						}
					}
				}
			}
		}
	}

	// Print the update in Slack format (no blank lines)
	fmt.Print("previous:\n")
	if len(yesterdayItems) > 0 {
		for _, item := range yesterdayItems {
			fmt.Printf("* %s\n", item)
		}
	} else {
		fmt.Print("* No work recorded\n")
	}

	fmt.Print("next:\n")
	if len(todayItems) > 0 {
		for _, item := range todayItems {
			fmt.Printf("* %s\n", item)
		}
	} else {
		fmt.Print("* No goals set\n")
	}

	return nil
}
