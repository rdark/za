package cmd

import (
	"fmt"
	"os"

	"github.com/rdark/za/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg     *config.Config
	version string
	commit  string
	date    string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "za",
	Short: "Za - Zettelkasten Augmentation tool",
	Long: `Za is a CLI tool for managing daily journal entries and standup notes.
It helps you extract work summaries, fix cross-reference links, and maintain
your zettelkasten-style knowledge base.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .za.yaml)")

	// Add version command
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Run: func(cmd *cobra.Command, args []string) {
		if version == "" {
			version = "dev"
		}
		fmt.Printf("za version %s\n", version)
		if commit != "" && commit != "none" {
			fmt.Printf("  commit: %s\n", commit)
		}
		if date != "" && date != "unknown" {
			fmt.Printf("  built: %s\n", date)
		}
	},
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	var err error
	cfg, err = config.Load(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
}

// GetConfig returns the loaded configuration
func GetConfig() *config.Config {
	return cfg
}

// SetVersionInfo sets the version information for the application
func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
}
