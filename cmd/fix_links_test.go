package cmd

import (
	"testing"

	"github.com/rdark/za/internal/notes"
)

func TestDetermineNoteType(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     notes.NoteType
		wantErr  bool
	}{
		{
			name:     "absolute journal path",
			filePath: "/path/to/journal/2025-10-27.md",
			want:     notes.NoteTypeJournal,
			wantErr:  false,
		},
		{
			name:     "relative journal path",
			filePath: "journal/2025-10-27.md",
			want:     notes.NoteTypeJournal,
			wantErr:  false,
		},
		{
			name:     "relative journal path with dot",
			filePath: "./journal/2025-10-27.md",
			want:     notes.NoteTypeJournal,
			wantErr:  false,
		},
		{
			name:     "absolute standup path",
			filePath: "/path/to/standup/2025-10-27.md",
			want:     notes.NoteTypeStandup,
			wantErr:  false,
		},
		{
			name:     "relative standup path",
			filePath: "standup/2025-10-27.md",
			want:     notes.NoteTypeStandup,
			wantErr:  false,
		},
		{
			name:     "windows journal path",
			filePath: "C:\\Users\\user\\journal\\2025-10-27.md",
			want:     notes.NoteTypeJournal,
			wantErr:  false,
		},
		{
			name:     "windows standup path",
			filePath: "C:\\Users\\user\\standup\\2025-10-27.md",
			want:     notes.NoteTypeStandup,
			wantErr:  false,
		},
		{
			name:     "case insensitive",
			filePath: "JOURNAL/2025-10-27.md",
			want:     notes.NoteTypeJournal,
			wantErr:  false,
		},
		{
			name:     "invalid path",
			filePath: "/path/to/notes/2025-10-27.md",
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := determineNoteType(tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("determineNoteType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("determineNoteType() = %v, want %v", got, tt.want)
			}
		})
	}
}
