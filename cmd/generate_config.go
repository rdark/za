package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	configOutput    string
	configForce     bool
	configWithNotes bool
	configMinimal   bool
)

var generateConfigCmd = &cobra.Command{
	Use:   "generate-config",
	Short: "Generate a sample configuration file",
	Long: `Generate a sample .za.yaml configuration file with sensible defaults.

By default, creates .za.yaml in the current directory with helpful comments
and examples. Use --minimal for a compact version without comments.

Examples:
  za generate-config                    # Create .za.yaml with comments
  za generate-config --minimal          # Create compact .za.yaml
  za generate-config --output my.yaml   # Create custom file
  za generate-config --force            # Overwrite existing config`,
	RunE: runGenerateConfig,
}

func init() {
	rootCmd.AddCommand(generateConfigCmd)
	generateConfigCmd.Flags().StringVarP(&configOutput, "output", "o", ".za.yaml", "Output file path")
	generateConfigCmd.Flags().BoolVar(&configForce, "force", false, "Overwrite existing config file")
	generateConfigCmd.Flags().BoolVar(&configWithNotes, "with-notes", true, "Include helpful comments (default: true)")
	generateConfigCmd.Flags().BoolVar(&configMinimal, "minimal", false, "Generate minimal config without comments")
}

func runGenerateConfig(cmd *cobra.Command, args []string) error {
	// Check if file exists
	if _, err := os.Stat(configOutput); err == nil && !configForce {
		return fmt.Errorf("config file already exists: %s (use --force to overwrite)", configOutput)
	}

	// Determine which template to use
	var configContent string
	if configMinimal {
		configContent = generateMinimalConfig()
	} else {
		configContent = generateFullConfig()
	}

	// Write config file
	if err := os.WriteFile(configOutput, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✓ Configuration file created: %s\n", configOutput)
	if !configMinimal {
		fmt.Println("\nNext steps:")
		fmt.Println("  1. Edit the file to customize paths and settings")
		fmt.Println("  2. Update 'dir' paths to point to your notes directories")
		fmt.Println("  3. Adjust 'work_done_sections' to match your note headings")
		fmt.Println("  4. (Optional) Set 'create.cmd' for journal/standup generation")
		fmt.Println("\nTest your config:")
		fmt.Printf("  za journal-work-done\n")
	}

	return nil
}

func generateFullConfig() string {
	return `# Za (Zettelkasten Augmentation) Configuration
#
# This file configures how za extracts work summaries, resolves links,
# and generates new notes in your zettelkasten.

# Journal Configuration
journal:
  # Directory containing journal entries (YYYY-MM-DD.md format)
  dir: ./journal

  # Section headings to extract for 'journal-work-done' command
  # za searches for these headings (case-insensitive) and extracts their content
  work_done_sections:
    - "work completed"
    - "worked on"

  # Text patterns to skip when extracting content (optional)
  skip_text: []

  # Synonyms for "previous day" links (used by fix-links command)
  # When you write [Yesterday](date), za knows to search backwards
  link_previous_titles:
    - "Yesterday"
    - "Previous"
    - "Last Week"

  # Synonyms for "next day" links
  link_next_titles:
    - "Tomorrow"
    - "Next"
    - "Next Week"

  # Command to create new journal entries (optional)
  # {date} placeholder will be replaced with YYYY-MM-DD format
  # Examples:
  #   cmd: "zk new --title 'Daily Log {date}' journal/"
  #   cmd: "~/scripts/create-journal.sh {date}"
  #   cmd: "touch journal/{date}.md && echo '---\ntitle: {date}\n---\n\n# Work Done\n\n' > journal/{date}.md"
  create:
    cmd: ""

# Standup Configuration
standup:
  # Directory containing standup notes (YYYY-MM-DD.md format)
  dir: ./standup

  # Single section heading to extract for 'standup-work-done' command
  # Unlike journal (which can have multiple sections), standup extracts one section
  work_done_section: "Worked on yesterday"

  # Text patterns to skip (optional)
  skip_text: []

  # Link synonyms (same as journal)
  link_previous_titles:
    - "Yesterday"
    - "Previous"
  link_next_titles:
    - "Tomorrow"
    - "Next"

  # Command to create new standup entries (optional)
  create:
    cmd: ""

# General Settings

# How many days to search backwards when looking for notes
# When a specific date doesn't exist, za searches backwards up to this many days
# Example: If you ask for 2025-01-09 (missing) and 2025-01-08 exists,
#          za will return 2025-01-08 if it's within the search window
search_window_days: 30
`
}

func generateMinimalConfig() string {
	return `journal:
  dir: ./journal
  work_done_sections:
    - "work completed"
    - "worked on"
  link_previous_titles: ["Yesterday", "Previous", "Last Week"]
  link_next_titles: ["Tomorrow", "Next", "Next Week"]
  create:
    cmd: ""

standup:
  dir: ./standup
  work_done_section: "Worked on yesterday"
  link_previous_titles: ["Yesterday", "Previous"]
  link_next_titles: ["Tomorrow", "Next"]
  create:
    cmd: ""

search_window_days: 30
`
}
