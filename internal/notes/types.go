package notes

// NoteType represents the type of note (journal, standup, etc.)
type NoteType string

const (
	// NoteTypeJournal represents a daily journal entry
	NoteTypeJournal NoteType = "journal"

	// NoteTypeStandup represents a daily standup note
	NoteTypeStandup NoteType = "standup"
)

// String returns the string representation of the note type
func (nt NoteType) String() string {
	return string(nt)
}

// IsValid checks if the note type is valid
func (nt NoteType) IsValid() bool {
	switch nt {
	case NoteTypeJournal, NoteTypeStandup:
		return true
	default:
		return false
	}
}
