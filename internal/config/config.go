package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the complete application configuration
type Config struct {
	Journal          JournalConfig `mapstructure:"journal"`
	Standup          StandupConfig `mapstructure:"standup"`
	GitHub           GitHubConfig  `mapstructure:"github"`
	SearchWindowDays int           `mapstructure:"search_window_days"`
	CompanyTag       string        `mapstructure:"company_tag"`
}

// JournalConfig contains configuration for journal notes
type JournalConfig struct {
	Dir                string        `mapstructure:"dir"`
	WorkDoneSections   []string      `mapstructure:"work_done_sections"`
	SkipText           []string      `mapstructure:"skip_text"`
	LinkPreviousTitles []string      `mapstructure:"link_previous_titles"`
	LinkNextTitles     []string      `mapstructure:"link_next_titles"`
	Create             CreateCommand `mapstructure:"create"`
}

// StandupConfig contains configuration for standup notes
type StandupConfig struct {
	Dir                string        `mapstructure:"dir"`
	WorkDoneSection    string        `mapstructure:"work_done_section"`
	SkipText           []string      `mapstructure:"skip_text"`
	LinkPreviousTitles []string      `mapstructure:"link_previous_titles"`
	LinkNextTitles     []string      `mapstructure:"link_next_titles"`
	Create             CreateCommand `mapstructure:"create"`
}

// CreateCommand contains the command to create new notes
type CreateCommand struct {
	Cmd string `mapstructure:"cmd"`
}

// GitHubConfig contains configuration for GitHub integration
type GitHubConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Org     string `mapstructure:"org"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Journal: JournalConfig{
			Dir:                "./journal",
			WorkDoneSections:   []string{"work completed", "worked on"},
			SkipText:           []string{},
			LinkPreviousTitles: []string{"Yesterday", "Previous", "Last Week"},
			LinkNextTitles:     []string{"Tomorrow", "Next", "Next Week"},
			Create:             CreateCommand{Cmd: ""},
		},
		Standup: StandupConfig{
			Dir:                "./standup",
			WorkDoneSection:    "Worked on yesterday",
			SkipText:           []string{},
			LinkPreviousTitles: []string{"Yesterday", "Previous", "Last Week"},
			LinkNextTitles:     []string{"Tomorrow", "Next", "Next Week"},
			Create:             CreateCommand{Cmd: ""},
		},
		GitHub: GitHubConfig{
			Enabled: false,
			Org:     "",
		},
		SearchWindowDays: 30,
		CompanyTag:       "acme",
	}
}

// Load loads configuration from file, environment variables, and defaults
// Precedence: CLI flags (passed separately) > env vars > config file > defaults
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Set environment variable prefix
	v.SetEnvPrefix("ZA")
	v.AutomaticEnv()

	// Load from config file if provided
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Look for .za.yaml in current directory and home directory
		v.SetConfigName(".za")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")

		// Add home directory
		if home, err := os.UserHomeDir(); err == nil {
			v.AddConfigPath(home)
		}
	}

	// Try to read config file (it's ok if it doesn't exist)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults
	}

	// Unmarshal into config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default values in viper
func setDefaults(v *viper.Viper) {
	defaults := DefaultConfig()

	v.SetDefault("journal.dir", defaults.Journal.Dir)
	v.SetDefault("journal.work_done_sections", defaults.Journal.WorkDoneSections)
	v.SetDefault("journal.skip_text", defaults.Journal.SkipText)
	v.SetDefault("journal.link_previous_titles", defaults.Journal.LinkPreviousTitles)
	v.SetDefault("journal.link_next_titles", defaults.Journal.LinkNextTitles)
	v.SetDefault("journal.create.cmd", defaults.Journal.Create.Cmd)

	v.SetDefault("standup.dir", defaults.Standup.Dir)
	v.SetDefault("standup.work_done_section", defaults.Standup.WorkDoneSection)
	v.SetDefault("standup.skip_text", defaults.Standup.SkipText)
	v.SetDefault("standup.link_previous_titles", defaults.Standup.LinkPreviousTitles)
	v.SetDefault("standup.link_next_titles", defaults.Standup.LinkNextTitles)
	v.SetDefault("standup.create.cmd", defaults.Standup.Create.Cmd)

	v.SetDefault("github.enabled", defaults.GitHub.Enabled)
	v.SetDefault("github.org", defaults.GitHub.Org)

	v.SetDefault("search_window_days", defaults.SearchWindowDays)
	v.SetDefault("company_tag", defaults.CompanyTag)
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Journal.Dir == "" {
		return fmt.Errorf("journal.dir is required")
	}
	if c.Standup.Dir == "" {
		return fmt.Errorf("standup.dir is required")
	}
	if c.SearchWindowDays <= 0 {
		return fmt.Errorf("search_window_days must be positive, got %d", c.SearchWindowDays)
	}
	if len(c.Journal.WorkDoneSections) == 0 {
		return fmt.Errorf("journal.work_done_sections must have at least one section")
	}
	if c.GitHub.Enabled && c.GitHub.Org == "" {
		return fmt.Errorf("github.org is required when github.enabled is true")
	}
	return nil
}

// ExpandPath expands relative paths to absolute paths
func (c *Config) ExpandPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	return filepath.Abs(path)
}

// JournalDir returns the absolute path to the journal directory
func (c *Config) JournalDir() (string, error) {
	return c.ExpandPath(c.Journal.Dir)
}

// StandupDir returns the absolute path to the standup directory
func (c *Config) StandupDir() (string, error) {
	return c.ExpandPath(c.Standup.Dir)
}
