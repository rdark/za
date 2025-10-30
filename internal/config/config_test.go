package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test journal defaults
	if cfg.Journal.Dir != "./journal" {
		t.Errorf("expected journal dir './journal', got %s", cfg.Journal.Dir)
	}
	if len(cfg.Journal.WorkDoneSections) != 2 {
		t.Errorf("expected 2 work done sections, got %d", len(cfg.Journal.WorkDoneSections))
	}
	if len(cfg.Journal.LinkPreviousTitles) != 3 {
		t.Errorf("expected 3 previous link titles, got %d", len(cfg.Journal.LinkPreviousTitles))
	}

	// Test standup defaults
	if cfg.Standup.Dir != "./standup" {
		t.Errorf("expected standup dir './standup', got %s", cfg.Standup.Dir)
	}
	if cfg.Standup.WorkDoneSection != "Worked on yesterday" {
		t.Errorf("expected work done section 'Worked on yesterday', got %s", cfg.Standup.WorkDoneSection)
	}

	// Test general defaults
	if cfg.SearchWindowDays != 30 {
		t.Errorf("expected search window 30 days, got %d", cfg.SearchWindowDays)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			cfg:     DefaultConfig(),
			wantErr: false,
		},
		{
			name: "missing journal dir",
			cfg: &Config{
				Journal: JournalConfig{
					Dir:              "",
					WorkDoneSections: []string{"work completed"},
				},
				Standup: StandupConfig{
					Dir: "./standup",
				},
				SearchWindowDays: 30,
			},
			wantErr: true,
			errMsg:  "journal.dir is required",
		},
		{
			name: "missing standup dir",
			cfg: &Config{
				Journal: JournalConfig{
					Dir:              "./journal",
					WorkDoneSections: []string{"work completed"},
				},
				Standup: StandupConfig{
					Dir: "",
				},
				SearchWindowDays: 30,
			},
			wantErr: true,
			errMsg:  "standup.dir is required",
		},
		{
			name: "invalid search window",
			cfg: &Config{
				Journal: JournalConfig{
					Dir:              "./journal",
					WorkDoneSections: []string{"work completed"},
				},
				Standup: StandupConfig{
					Dir: "./standup",
				},
				SearchWindowDays: 0,
			},
			wantErr: true,
			errMsg:  "search_window_days must be positive",
		},
		{
			name: "missing work done sections",
			cfg: &Config{
				Journal: JournalConfig{
					Dir:              "./journal",
					WorkDoneSections: []string{},
				},
				Standup: StandupConfig{
					Dir: "./standup",
				},
				SearchWindowDays: 30,
			},
			wantErr: true,
			errMsg:  "journal.work_done_sections must have at least one section",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				// Check if error message contains the expected substring
				if len(tt.errMsg) > 0 && err.Error()[:len(tt.errMsg)] != tt.errMsg {
					t.Errorf("Validate() error message = %v, want %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".za.yaml")

	configContent := `
journal:
  dir: /tmp/test-journal
  work_done_sections:
    - "completed work"
    - "worked on tasks"
  skip_text:
    - "skip this"
  link_previous_titles:
    - "Yesterday"
    - "Previous"
  link_next_titles:
    - "Tomorrow"
    - "Next"

standup:
  dir: /tmp/test-standup
  work_done_section: "Yesterday's work"
  link_previous_titles:
    - "Previous Day"
  link_next_titles:
    - "Next Day"

search_window_days: 45
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify loaded values
	if cfg.Journal.Dir != "/tmp/test-journal" {
		t.Errorf("expected journal dir '/tmp/test-journal', got %s", cfg.Journal.Dir)
	}
	if len(cfg.Journal.WorkDoneSections) != 2 {
		t.Errorf("expected 2 work done sections, got %d", len(cfg.Journal.WorkDoneSections))
	}
	if cfg.Journal.WorkDoneSections[0] != "completed work" {
		t.Errorf("expected first section 'completed work', got %s", cfg.Journal.WorkDoneSections[0])
	}
	if cfg.Standup.WorkDoneSection != "Yesterday's work" {
		t.Errorf("expected standup work section 'Yesterday's work', got %s", cfg.Standup.WorkDoneSection)
	}
	if cfg.SearchWindowDays != 45 {
		t.Errorf("expected search window 45, got %d", cfg.SearchWindowDays)
	}
	if len(cfg.Journal.SkipText) != 1 || cfg.Journal.SkipText[0] != "skip this" {
		t.Errorf("expected skip_text ['skip this'], got %v", cfg.Journal.SkipText)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	// Create a temporary directory without any config file
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()

	// Change to temp directory so no config file is found
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Load config without a file (should use defaults)
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() should succeed with defaults, got error: %v", err)
	}

	// Should match defaults
	defaults := DefaultConfig()
	if cfg.Journal.Dir != defaults.Journal.Dir {
		t.Errorf("expected default journal dir, got %s", cfg.Journal.Dir)
	}
	if cfg.SearchWindowDays != defaults.SearchWindowDays {
		t.Errorf("expected default search window, got %d", cfg.SearchWindowDays)
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".za.yaml")

	// Write invalid YAML
	invalidYAML := `
journal:
  dir: /tmp/journal
  invalid yaml here [[[
`
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Load() should fail with invalid YAML")
	}
}

func TestLoadConfigInvalidValues(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".za.yaml")

	// Config with invalid values (negative search window)
	configContent := `
journal:
  dir: /tmp/journal
  work_done_sections:
    - "work"
standup:
  dir: /tmp/standup
search_window_days: -5
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Load() should fail with negative search_window_days")
	}
}

func TestExpandPath(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		name    string
		path    string
		wantAbs bool
	}{
		{
			name:    "relative path",
			path:    "./test",
			wantAbs: true,
		},
		{
			name:    "absolute path",
			path:    "/tmp/test",
			wantAbs: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expanded, err := cfg.ExpandPath(tt.path)
			if err != nil {
				t.Errorf("ExpandPath() error = %v", err)
				return
			}
			if tt.wantAbs && !filepath.IsAbs(expanded) {
				t.Errorf("ExpandPath() = %v, want absolute path", expanded)
			}
		})
	}
}

func TestJournalDir(t *testing.T) {
	cfg := DefaultConfig()
	dir, err := cfg.JournalDir()
	if err != nil {
		t.Errorf("JournalDir() error = %v", err)
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("JournalDir() = %v, want absolute path", dir)
	}
}

func TestStandupDir(t *testing.T) {
	cfg := DefaultConfig()
	dir, err := cfg.StandupDir()
	if err != nil {
		t.Errorf("StandupDir() error = %v", err)
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("StandupDir() = %v, want absolute path", dir)
	}
}
