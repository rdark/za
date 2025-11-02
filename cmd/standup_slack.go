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

This command reads from the standup file and extracts:
- Work completed yesterday from "Worked on Yesterday" section
- Planned work for today from "Working on Today" section

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

	standupDir, err := cfg.StandupDir()
	if err != nil {
		return fmt.Errorf("failed to get standup directory: %w", err)
	}

	// Find today's standup
	standupPath, err := notes.FindNoteByDate(targetDate, notes.NoteTypeStandup, standupDir, cfg.SearchWindowDays)
	if err != nil {
		return fmt.Errorf("no standup found for %s: %w", targetDate.Format(notes.DateFormat), err)
	}

	// Verify this is actually today's standup
	foundDate, err := notes.ParseDateFromFilename(standupPath)
	if err != nil {
		return fmt.Errorf("failed to parse date from standup filename: %w", err)
	}

	targetY, targetM, targetD := targetDate.Date()
	foundY, foundM, foundD := foundDate.Date()
	if targetY != foundY || targetM != foundM || targetD != foundD {
		return fmt.Errorf("no standup found for exact date %s (found %s)",
			targetDate.Format(notes.DateFormat), foundDate.Format(notes.DateFormat))
	}

	// Parse standup file
	parser := markdown.NewParser()
	standupDoc, err := parser.ParseFile(standupPath)
	if err != nil {
		return fmt.Errorf("failed to parse standup file: %w", err)
	}

	// Extract yesterday's work from "Worked on Yesterday" section
	var yesterdayItems []string
	yesterdaySection := standupDoc.FindSectionByHeading(cfg.Standup.WorkDoneSection)
	if yesterdaySection != nil && strings.TrimSpace(yesterdaySection.Content) != "" {
		lines := strings.Split(yesterdaySection.Content, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			// Skip navigation links (Yesterday, Today, Tomorrow, Standup, Daily)
			if strings.HasPrefix(trimmed, "* [Yesterday") || strings.HasPrefix(trimmed, "* [Today") ||
				strings.HasPrefix(trimmed, "* [Tomorrow") || strings.HasPrefix(trimmed, "* [Standup") ||
				strings.HasPrefix(trimmed, "* [Daily") ||
				strings.HasPrefix(trimmed, "- [Yesterday") || strings.HasPrefix(trimmed, "- [Today") ||
				strings.HasPrefix(trimmed, "- [Tomorrow") || strings.HasPrefix(trimmed, "- [Standup") ||
				strings.HasPrefix(trimmed, "- [Daily") {
				continue
			}
			// Extract bullet points
			var item string
			if strings.HasPrefix(trimmed, "* ") {
				item = strings.TrimSpace(strings.TrimPrefix(trimmed, "* "))
			} else if strings.HasPrefix(trimmed, "- ") {
				item = strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			}
			if item != "" {
				yesterdayItems = append(yesterdayItems, item)
			}
		}
	}

	// Extract today's goals from "Working on Today" section
	var todayItems []string
	todaySection := standupDoc.FindSectionByHeading("Working on Today")
	if todaySection != nil && strings.TrimSpace(todaySection.Content) != "" {
		lines := strings.Split(todaySection.Content, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			// Skip navigation links (Yesterday, Today, Tomorrow, Standup, Daily)
			if strings.HasPrefix(trimmed, "* [Yesterday") || strings.HasPrefix(trimmed, "* [Today") ||
				strings.HasPrefix(trimmed, "* [Tomorrow") || strings.HasPrefix(trimmed, "* [Standup") ||
				strings.HasPrefix(trimmed, "* [Daily") ||
				strings.HasPrefix(trimmed, "- [Yesterday") || strings.HasPrefix(trimmed, "- [Today") ||
				strings.HasPrefix(trimmed, "- [Tomorrow") || strings.HasPrefix(trimmed, "- [Standup") ||
				strings.HasPrefix(trimmed, "- [Daily") {
				continue
			}
			// Extract bullet points
			var item string
			if strings.HasPrefix(trimmed, "* ") {
				item = strings.TrimSpace(strings.TrimPrefix(trimmed, "* "))
			} else if strings.HasPrefix(trimmed, "- ") {
				item = strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			}
			if item != "" {
				todayItems = append(todayItems, item)
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
