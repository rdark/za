package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rdark/za/internal/links"
	"github.com/rdark/za/internal/markdown"
	"github.com/rdark/za/internal/notes"
	"github.com/rdark/za/internal/util"
	"github.com/spf13/cobra"
)

var (
	skipWorkExtraction bool
)

var generateJournalCmd = &cobra.Command{
	Use:   "generate-journal [date]",
	Short: "Generate a new journal entry",
	Long: `Generate a new journal entry using the configured create command.

This command executes the journal create command specified in the configuration.
After creation, it automatically fixes any relative date links in the new file.

Examples:
  za generate-journal                    # Generate today's journal
  za generate-journal 2025-01-15        # Generate journal for specific date`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGenerateJournal,
}

var generateStandupCmd = &cobra.Command{
	Use:   "generate-standup [date]",
	Short: "Generate a new standup entry",
	Long: `Generate a new standup entry using the configured create command.

This command executes the standup create command specified in the configuration.
By default, it extracts work from the previous day's journal and populates the standup.
After creation, it automatically fixes any relative date links in the new file.

Examples:
  za generate-standup                    # Generate today's standup with yesterday's work
  za generate-standup 2025-01-15        # Generate standup for specific date
  za generate-standup --no-work         # Generate without extracting work from journal`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGenerateStandup,
}

func init() {
	rootCmd.AddCommand(generateJournalCmd)
	rootCmd.AddCommand(generateStandupCmd)

	generateStandupCmd.Flags().BoolVar(&skipWorkExtraction, "no-work", false, "Skip populating with work from previous day's journal")
}

func runGenerateJournal(cmd *cobra.Command, args []string) error {
	// Parse target date
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

	// Check if create command is configured
	if cfg.Journal.Create.Cmd == "" {
		return fmt.Errorf("journal.create.cmd is not configured in .za.yaml")
	}

	// Get journal directory
	journalDir, err := cfg.JournalDir()
	if err != nil {
		return fmt.Errorf("failed to get journal directory: %w", err)
	}

	// Build expected file path
	dateStr := targetDate.Format(notes.DateFormat)
	expectedPath := filepath.Join(journalDir, dateStr+".md")

	// Check if file already exists
	if _, err := os.Stat(expectedPath); err == nil {
		return fmt.Errorf("journal entry already exists: %s", expectedPath)
	}

	fmt.Printf("Generating journal entry for %s...\n", dateStr)

	// Replace {date} placeholder in command
	createCmd := strings.ReplaceAll(cfg.Journal.Create.Cmd, "{date}", dateStr)

	// Execute create command
	result := util.ExecuteShellCommand(createCmd, util.DefaultTimeout)

	if result.Error != nil {
		fmt.Fprintf(os.Stderr, "Failed to execute create command:\n")
		fmt.Fprintf(os.Stderr, "Command: %s\n", createCmd)
		fmt.Fprintf(os.Stderr, "Error: %v\n", result.Error)
		if result.Stderr != "" {
			fmt.Fprintf(os.Stderr, "Stderr: %s\n", result.Stderr)
		}
		return fmt.Errorf("create command failed with exit code %d", result.ExitCode)
	}

	// Verify file was created
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		// Try to find any newly created file in the journal directory
		files, err := filepath.Glob(filepath.Join(journalDir, dateStr+"*.md"))
		if err != nil {
			return fmt.Errorf("failed to search for created file: %w", err)
		}
		if len(files) > 0 {
			expectedPath = files[0]
			fmt.Printf("✓ Journal entry created: %s\n", expectedPath)
		} else {
			fmt.Printf("⚠ Create command succeeded but file not found at expected path: %s\n", expectedPath)
			if result.Stdout != "" {
				fmt.Printf("Command output: %s\n", result.Stdout)
			}
			return nil
		}
	} else {
		fmt.Printf("✓ Journal entry created: %s\n", expectedPath)
	}

	// Add company tag if it's a weekday and tag is configured
	if cfg.CompanyTag != "" && util.IsWeekday(targetDate) {
		fmt.Println("\nAdding company tag...")
		companyTag := fmt.Sprintf("company:%s", cfg.CompanyTag)
		if added, err := markdown.AddTagToFile(expectedPath, companyTag); err != nil {
			fmt.Fprintf(os.Stderr, "⚠ Failed to add company tag: %v\n", err)
		} else if added {
			fmt.Printf("✓ Added tag: %s\n", companyTag)
		}
	}

	// Populate goals from previous journal
	fmt.Println("\nPopulating goals from previous journal...")
	if err := populateJournalGoals(targetDate, expectedPath); err != nil {
		fmt.Fprintf(os.Stderr, "⚠ Failed to populate goals: %v\n", err)
		// Don't fail the command if goals population fails
	}

	// Automatically fix links in the created file
	fmt.Println("\nFixing links...")
	if err := fixLinksInFile(expectedPath); err != nil {
		fmt.Fprintf(os.Stderr, "⚠ Failed to fix links: %v\n", err)
		// Don't fail the command if link fixing fails
	}

	return nil
}

func runGenerateStandup(cmd *cobra.Command, args []string) error {
	// Parse target date
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

	// Check if create command is configured
	if cfg.Standup.Create.Cmd == "" {
		return fmt.Errorf("standup.create.cmd is not configured in .za.yaml")
	}

	// Get standup directory
	standupDir, err := cfg.StandupDir()
	if err != nil {
		return fmt.Errorf("failed to get standup directory: %w", err)
	}

	// Build expected file path
	dateStr := targetDate.Format(notes.DateFormat)
	expectedPath := filepath.Join(standupDir, dateStr+".md")

	// Check if file already exists
	if _, err := os.Stat(expectedPath); err == nil {
		return fmt.Errorf("standup entry already exists: %s", expectedPath)
	}

	fmt.Printf("Generating standup entry for %s...\n", dateStr)

	// Replace {date} placeholder in command
	createCmd := strings.ReplaceAll(cfg.Standup.Create.Cmd, "{date}", dateStr)

	// Execute create command
	result := util.ExecuteShellCommand(createCmd, util.DefaultTimeout)

	if result.Error != nil {
		fmt.Fprintf(os.Stderr, "Failed to execute create command:\n")
		fmt.Fprintf(os.Stderr, "Command: %s\n", createCmd)
		fmt.Fprintf(os.Stderr, "Error: %v\n", result.Error)
		if result.Stderr != "" {
			fmt.Fprintf(os.Stderr, "Stderr: %s\n", result.Stderr)
		}
		return fmt.Errorf("create command failed with exit code %d", result.ExitCode)
	}

	// Verify file was created
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		// Try to find any newly created file in the standup directory
		files, err := filepath.Glob(filepath.Join(standupDir, dateStr+"*.md"))
		if err != nil {
			return fmt.Errorf("failed to search for created file: %w", err)
		}
		if len(files) > 0 {
			expectedPath = files[0]
			fmt.Printf("✓ Standup entry created: %s\n", expectedPath)
		} else {
			fmt.Printf("⚠ Create command succeeded but file not found at expected path: %s\n", expectedPath)
			if result.Stdout != "" {
				fmt.Printf("Command output: %s\n", result.Stdout)
			}
			return nil
		}
	} else {
		fmt.Printf("✓ Standup entry created: %s\n", expectedPath)
	}

	// Add company tag if it's a weekday and tag is configured
	if cfg.CompanyTag != "" && util.IsWeekday(targetDate) {
		fmt.Println("\nAdding company tag...")
		companyTag := fmt.Sprintf("company:%s", cfg.CompanyTag)
		if added, err := markdown.AddTagToFile(expectedPath, companyTag); err != nil {
			fmt.Fprintf(os.Stderr, "⚠ Failed to add company tag: %v\n", err)
		} else if added {
			fmt.Printf("✓ Added tag: %s\n", companyTag)
		}
	}

	// Extract work from previous journal by default
	if !skipWorkExtraction {
		fmt.Println("\nExtracting work from previous journal...")
		if err := populateStandupWithWork(targetDate, expectedPath); err != nil {
			fmt.Fprintf(os.Stderr, "⚠ Failed to extract work: %v\n", err)
			// Don't fail the command if work extraction fails
		}
	}

	// Automatically fix links in the created file
	fmt.Println("\nFixing links...")
	if err := fixLinksInFile(expectedPath); err != nil {
		fmt.Fprintf(os.Stderr, "⚠ Failed to fix links: %v\n", err)
		// Don't fail the command if link fixing fails
	}

	return nil
}

// populateStandupWithWork extracts work from previous day's journal and today's goals,
// inserting them into the appropriate standup sections
func populateStandupWithWork(standupDate time.Time, standupPath string) error {
	journalDir, err := cfg.JournalDir()
	if err != nil {
		return err
	}

	// Find previous day's journal for "Worked on Yesterday" section
	previousDate := standupDate.AddDate(0, 0, -1)
	prevJournalPath, err := notes.FindNoteByDate(previousDate, notes.NoteTypeJournal, journalDir, cfg.SearchWindowDays)
	if err != nil {
		return fmt.Errorf("could not find previous journal: %w", err)
	}

	fmt.Printf("Found previous journal: %s\n", prevJournalPath)

	// Parse previous journal
	parser := markdown.NewParser()
	prevDoc, err := parser.ParseFile(prevJournalPath)
	if err != nil {
		return fmt.Errorf("failed to parse previous journal: %w", err)
	}

	// Extract work sections from previous journal
	workSections := prevDoc.FindSectionsByHeadings(cfg.Journal.WorkDoneSections)

	// Extract completed goals from previous journal's "Goals of the Day"
	var completedGoals []string
	prevGoalsSection := prevDoc.FindSectionByHeading("Goals of the Day")
	if prevGoalsSection != nil && strings.TrimSpace(prevGoalsSection.Content) != "" {
		items := markdown.ParseGoalItems(prevGoalsSection.Content)
		for _, item := range items {
			// Only include completed checkbox items
			if item.HasCheckbox && item.Checked {
				completedGoals = append(completedGoals, item.Text)
			}
		}
	}

	// Build content for "Worked on Yesterday" section
	var yesterdayContent strings.Builder
	if len(completedGoals) > 0 {
		fmt.Printf("Adding %d completed goal(s) from yesterday\n", len(completedGoals))
		for _, goal := range completedGoals {
			yesterdayContent.WriteString(fmt.Sprintf("* %s\n", goal))
		}
		// No extra newline - keep goals and work items together
	}
	for _, section := range workSections {
		sectionContent := strings.TrimSpace(section.Content)
		if sectionContent != "" {
			yesterdayContent.WriteString(sectionContent)
			yesterdayContent.WriteString("\n\n")
		}
	}

	// Find today's journal for "Working on Today" section
	var todayGoals []string
	todayJournalPath, err := notes.FindNoteByDate(standupDate, notes.NoteTypeJournal, journalDir, cfg.SearchWindowDays)
	if err == nil {
		// Verify this is actually today's journal, not a fallback to an earlier date
		foundDate, err := notes.ParseDateFromFilename(todayJournalPath)
		if err == nil {
			// Compare just the date parts (year, month, day) not the full timestamp
			standupY, standupM, standupD := standupDate.Date()
			foundY, foundM, foundD := foundDate.Date()
			if standupY == foundY && standupM == foundM && standupD == foundD {
				fmt.Printf("Found today's journal: %s\n", todayJournalPath)

				todayDoc, err := parser.ParseFile(todayJournalPath)
				if err == nil {
					todayGoalsSection := todayDoc.FindSectionByHeading("Goals of the Day")
					if todayGoalsSection != nil && strings.TrimSpace(todayGoalsSection.Content) != "" {
						items := markdown.ParseGoalItems(todayGoalsSection.Content)
						// Include all goals (completed and uncompleted)
						for _, item := range items {
							if item.HasCheckbox || item.Text != "" {
								todayGoals = append(todayGoals, item.Text)
							}
						}
					}
				}
			} else {
				fmt.Println("No today's journal found yet (found fallback from earlier date)")
			}
		}
	} else {
		fmt.Println("No today's journal found yet")
	}

	// Build content for "Working on Today" section
	var todayContent strings.Builder
	if len(todayGoals) > 0 {
		fmt.Printf("Adding %d goal(s) for today\n", len(todayGoals))
		for _, goal := range todayGoals {
			todayContent.WriteString(fmt.Sprintf("* %s\n", goal))
		}
		todayContent.WriteString("\n")
	}

	// Read current standup content
	standupContent, err := os.ReadFile(standupPath)
	if err != nil {
		return fmt.Errorf("failed to read standup file: %w", err)
	}

	// Insert content into standup sections
	newContent := string(standupContent)

	if yesterdayContent.Len() > 0 {
		// Add leading newline for spacing after existing content (like links)
		content := "\n" + yesterdayContent.String()
		newContent, err = insertIntoStandupSection(newContent, cfg.Standup.WorkDoneSection, content)
		if err != nil {
			return fmt.Errorf("failed to insert yesterday's work: %w", err)
		}
	}

	if todayContent.Len() > 0 {
		// Add leading newline for spacing after existing content (like links)
		content := "\n" + todayContent.String()
		newContent, err = insertIntoStandupSection(newContent, "Working on Today", content)
		if err != nil {
			return fmt.Errorf("failed to insert today's goals: %w", err)
		}
	}

	// Write updated content back to file
	if err := os.WriteFile(standupPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write standup file: %w", err)
	}

	fmt.Printf("✓ Populated standup with work from %s\n", filepath.Base(prevJournalPath))
	return nil
}

// fixLinksInFile fixes all relative date links in the given file
func fixLinksInFile(filePath string) error {
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

	// Extract and fix links
	allLinks := doc.ExtractLinks()
	if len(allLinks) == 0 {
		return nil // No links to fix
	}

	// Classify, resolve, and filter links that need fixing
	needsUpdate, err := classifyAndResolveLinks(allLinks, fileDate, noteType)
	if err != nil {
		return err
	}

	if len(needsUpdate) == 0 {
		return nil // All links are correct
	}

	fmt.Printf("Fixing %d links...\n", len(needsUpdate))

	// Apply changes
	newContent, err := applyLinkFixes(doc, needsUpdate)
	if err != nil {
		return fmt.Errorf("failed to apply link fixes: %w", err)
	}

	// Write back to file
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("✓ Fixed %d links in %s\n", len(needsUpdate), filepath.Base(filePath))
	return nil
}

// hasGoalContent checks if a section has actual goal items (not just comments or whitespace)
func hasGoalContent(sectionContent string) bool {
	lines := strings.Split(sectionContent, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines and HTML comments
		if trimmed == "" || strings.HasPrefix(trimmed, "<!--") {
			continue
		}
		// If we find any non-comment, non-empty line, the section has content
		return true
	}
	return false
}

// populateJournalGoals populates goals sections from the previous journal entry
func populateJournalGoals(currentDate time.Time, journalPath string) error {
	// Find previous journal
	previousDate := currentDate.AddDate(0, 0, -1)
	journalDir, err := cfg.JournalDir()
	if err != nil {
		return err
	}

	prevJournalPath, err := notes.FindNoteByDate(previousDate, notes.NoteTypeJournal, journalDir, cfg.SearchWindowDays)
	if err != nil {
		// No previous journal found - this is fine
		fmt.Println("No previous journal found to copy goals from")
		return nil
	}

	fmt.Printf("Found previous journal: %s\n", filepath.Base(prevJournalPath))

	// Parse previous journal
	parser := markdown.NewParser()
	prevDoc, err := parser.ParseFile(prevJournalPath)
	if err != nil {
		return fmt.Errorf("failed to parse previous journal: %w", err)
	}

	// Parse the actual date from the previous journal filename
	prevDate, err := notes.ParseDateFromFilename(prevJournalPath)
	if err != nil {
		return fmt.Errorf("failed to parse date from previous journal: %w", err)
	}

	// Read current journal content
	currentContent, err := os.ReadFile(journalPath)
	if err != nil {
		return fmt.Errorf("failed to read current journal: %w", err)
	}

	content := string(currentContent)

	// Parse current document to check for existing goals sections
	currentDoc, err := parser.ParseFile(journalPath)
	if err != nil {
		return fmt.Errorf("failed to parse current journal: %w", err)
	}

	var goalsToAdd strings.Builder
	sectionsAdded := false

	// 1. Copy "Goals of the Week" if same week (FIRST)
	if util.IsSameWeek(prevDate, currentDate) {
		weekGoalsSection := prevDoc.FindSectionByHeading("Goals of the Week")
		if weekGoalsSection != nil && strings.TrimSpace(weekGoalsSection.Content) != "" {
			// Check if current journal has this section with content
			currentWeekSection := currentDoc.FindSectionByHeading("Goals of the Week")
			shouldAdd := currentWeekSection == nil || !hasGoalContent(currentWeekSection.Content)

			if shouldAdd {
				fmt.Println("Copying Goals of the Week (same week)")
				goalsToAdd.WriteString("## Goals of the Week\n\n")
				goalsToAdd.WriteString(strings.TrimSpace(weekGoalsSection.Content))
				goalsToAdd.WriteString("\n\n")
				sectionsAdded = true
			}
		}
	}

	// 2. Copy unfinished "Goals of the Day" items (SECOND)
	// Always add this section, even if empty
	currentDaySection := currentDoc.FindSectionByHeading("Goals of the Day")
	shouldAddDayGoals := currentDaySection == nil || !hasGoalContent(currentDaySection.Content)

	if shouldAddDayGoals {
		dayGoalsSection := prevDoc.FindSectionByHeading("Goals of the Day")
		var unfinishedItems []markdown.GoalItem

		if dayGoalsSection != nil && strings.TrimSpace(dayGoalsSection.Content) != "" {
			// Parse both checkbox items and plain bullet points
			items := markdown.ParseGoalItems(dayGoalsSection.Content)
			unfinishedItems = markdown.FilterUnfinishedGoals(items)
		}

		if len(unfinishedItems) > 0 {
			fmt.Printf("Copying %d unfinished goal(s) from yesterday\n", len(unfinishedItems))
			formattedItems := markdown.FormatGoalItems(unfinishedItems)
			goalsToAdd.WriteString("## Goals of the Day\n\n")
			goalsToAdd.WriteString(formattedItems)
			goalsToAdd.WriteString("\n\n")
		} else {
			fmt.Println("Adding empty Goals of the Day section")
			goalsToAdd.WriteString("## Goals of the Day\n\n")
		}
		sectionsAdded = true
	}

	// Insert goals sections after Daily Log heading if any were added
	if sectionsAdded {
		newContent, err := insertAfterDailyLogSection(content, goalsToAdd.String())
		if err != nil {
			return fmt.Errorf("failed to insert goals: %w", err)
		}

		// Write updated content back to file
		if err := os.WriteFile(journalPath, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to write journal file: %w", err)
		}

		fmt.Println("✓ Goals populated successfully")
	} else {
		fmt.Println("No goals to populate")
	}

	return nil
}

// insertAfterDailyLogSection inserts content after the Daily Log h1 section,
// removing any empty Goals sections that already exist
func insertAfterDailyLogSection(fileContent, insertContent string) (string, error) {
	// Check which sections we're inserting
	insertingGoalsOfDay := strings.Contains(insertContent, "## Goals of the Day")
	insertingGoalsOfWeek := strings.Contains(insertContent, "## Goals of the Week")
	lines := strings.Split(fileContent, "\n")

	// Find the first h1 heading (Daily Log)
	h1Index := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") && !strings.HasPrefix(trimmed, "## ") {
			h1Index = i
			break
		}
	}

	if h1Index == -1 {
		// No h1 heading found, insert at the beginning after frontmatter
		return insertAfterFrontmatter(fileContent, insertContent)
	}

	// Find where to insert: after the h1 and any links that follow
	insertIndex := h1Index + 1

	// Skip blank lines
	for insertIndex < len(lines) && strings.TrimSpace(lines[insertIndex]) == "" {
		insertIndex++
	}

	// Skip bullet list of links (*, [Yesterday], [Tomorrow], [Standup], etc.)
	for insertIndex < len(lines) {
		trimmed := strings.TrimSpace(lines[insertIndex])
		if trimmed == "" || strings.HasPrefix(trimmed, "* [") || strings.HasPrefix(trimmed, "- [") {
			insertIndex++
		} else {
			break
		}
	}

	// Remove any existing empty Goals sections that we're about to replace
	filteredLines := make([]string, 0, len(lines))
	filteredLines = append(filteredLines, lines[:insertIndex]...)

	// Process remaining lines, removing empty Goals sections
	i := insertIndex
	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])

		// Check if this is a Goals heading
		if trimmed == "## Goals of the Week" || trimmed == "## Goals of the Day" {
			// Find the extent of this section (until next heading or end of file)
			sectionStart := i
			sectionHeading := trimmed
			i++

			// Collect content until next heading
			var sectionContent []string
			for i < len(lines) {
				lineTrimmed := strings.TrimSpace(lines[i])
				if strings.HasPrefix(lineTrimmed, "#") {
					// Hit next heading
					break
				}
				sectionContent = append(sectionContent, lines[i])
				i++
			}

			// Decide whether to keep this section:
			// - If we're inserting a new version of this section, remove empty existing ones
			// - If we're not inserting this section, preserve it even if empty
			// - Always preserve sections with actual content
			hasContent := hasGoalContent(strings.Join(sectionContent, "\n"))
			shouldKeep := hasContent

			if !shouldKeep {
				// Check if we should preserve this empty section
				if sectionHeading == "## Goals of the Day" && !insertingGoalsOfDay {
					shouldKeep = true
				} else if sectionHeading == "## Goals of the Week" && !insertingGoalsOfWeek {
					shouldKeep = true
				}
			}

			if shouldKeep {
				// Keep the section
				filteredLines = append(filteredLines, lines[sectionStart])
				filteredLines = append(filteredLines, sectionContent...)
			}
		} else {
			filteredLines = append(filteredLines, lines[i])
			i++
		}
	}

	// Build result
	var result strings.Builder

	// Write everything up to insertion point
	for i := 0; i < len(filteredLines) && i < insertIndex; i++ {
		result.WriteString(filteredLines[i])
		result.WriteString("\n")
	}

	// Check if we need to add a blank line before inserted content
	// (only add if the last line written wasn't already blank)
	if insertIndex > 0 && strings.TrimSpace(filteredLines[insertIndex-1]) != "" {
		result.WriteString("\n")
	}

	// Write inserted content
	result.WriteString(insertContent)

	// Write rest of filtered lines
	for i := insertIndex; i < len(filteredLines); i++ {
		result.WriteString(filteredLines[i])
		if i < len(filteredLines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String(), nil
}

// insertAfterFrontmatter inserts content after the frontmatter section
func insertAfterFrontmatter(fileContent, insertContent string) (string, error) {
	lines := strings.Split(fileContent, "\n")

	// Find the end of frontmatter (second occurrence of ---)
	frontmatterEnd := -1
	dashCount := 0
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			dashCount++
			if dashCount == 2 {
				frontmatterEnd = i
				break
			}
		}
	}

	if frontmatterEnd == -1 {
		// No frontmatter, insert at the beginning
		return insertContent + "\n" + fileContent, nil
	}

	// Insert after frontmatter
	var result strings.Builder

	// Write frontmatter
	for i := 0; i <= frontmatterEnd; i++ {
		result.WriteString(lines[i])
		result.WriteString("\n")
	}

	// Write inserted content
	result.WriteString(insertContent)

	// Write rest of file
	for i := frontmatterEnd + 1; i < len(lines); i++ {
		result.WriteString(lines[i])
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String(), nil
}

// insertIntoStandupSection inserts content into a specific section of a standup file
func insertIntoStandupSection(fileContent, sectionHeading, insertContent string) (string, error) {
	lines := strings.Split(fileContent, "\n")

	// Find the section heading (case-insensitive, supports both # and ## headings)
	sectionIndex := -1
	sectionLevel := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Check for h1 or h2 heading
		if strings.HasPrefix(trimmed, "##") {
			headingText := strings.TrimSpace(strings.TrimPrefix(trimmed, "##"))
			if strings.EqualFold(headingText, sectionHeading) {
				sectionIndex = i
				sectionLevel = 2
				break
			}
		} else if strings.HasPrefix(trimmed, "#") {
			headingText := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
			if strings.EqualFold(headingText, sectionHeading) {
				sectionIndex = i
				sectionLevel = 1
				break
			}
		}
	}

	if sectionIndex == -1 {
		return fileContent, fmt.Errorf("section '%s' not found", sectionHeading)
	}

	// Find where to insert: after the heading and any existing content, before next heading
	insertIndex := sectionIndex + 1

	// Skip blank lines after heading
	for insertIndex < len(lines) && strings.TrimSpace(lines[insertIndex]) == "" {
		insertIndex++
	}

	// Skip existing content until we hit another heading of same or higher level
	for insertIndex < len(lines) {
		trimmed := strings.TrimSpace(lines[insertIndex])

		// Check if this is a heading
		if strings.HasPrefix(trimmed, "#") {
			// Count heading level
			level := 0
			for _, ch := range trimmed {
				if ch == '#' {
					level++
				} else {
					break
				}
			}

			// If same or higher level heading, insert before it
			if level <= sectionLevel {
				// Back up to skip trailing blank lines
				for insertIndex > sectionIndex+1 && strings.TrimSpace(lines[insertIndex-1]) == "" {
					insertIndex--
				}
				break
			}
		}

		insertIndex++
	}

	// Build result
	var result strings.Builder

	// Write everything up to insertion point
	for i := 0; i < insertIndex; i++ {
		result.WriteString(lines[i])
		result.WriteString("\n")
	}

	// Write inserted content
	result.WriteString(insertContent)

	// Write rest of file
	for i := insertIndex; i < len(lines); i++ {
		result.WriteString(lines[i])
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String(), nil
}

// classifyAndResolveLinks classifies and resolves links, returning only those that need updating
func classifyAndResolveLinks(allLinks []markdown.Link, fileDate time.Time, noteType notes.NoteType) ([]links.ResolvedLink, error) {
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
		return nil, nil
	}

	// Resolve links
	resolver := links.NewResolver(cfg, fileDate, noteType)
	resolved := resolver.ResolveAll(fixable)

	// Filter to links that need updating
	needsUpdate := links.FilterNeedsUpdate(resolved)

	return needsUpdate, nil
}
