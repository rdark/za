package links

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/rdark/za/internal/config"
	"github.com/rdark/za/internal/notes"
)

// ResolvedLink represents a link with its resolved target
type ResolvedLink struct {
	// Classified is the classified link
	Classified ClassifiedLink

	// ResolvedPath is the actual file path the link should point to (if resolved)
	ResolvedPath string

	// ResolvedDate is the date of the resolved note
	ResolvedDate time.Time

	// Error is set if the link couldn't be resolved
	Error error

	// NeedsUpdate is true if the link destination needs to be updated
	NeedsUpdate bool

	// SuggestedDestination is the suggested new destination for the link
	SuggestedDestination string
}

// Resolver resolves links to actual file paths
type Resolver struct {
	cfg             *config.Config
	currentDate     time.Time
	currentNoteType notes.NoteType
}

// NewResolver creates a new link resolver
// currentDate is the date of the current note being processed
// currentNoteType is the type of the current note (journal or standup)
func NewResolver(cfg *config.Config, currentDate time.Time, currentNoteType notes.NoteType) *Resolver {
	return &Resolver{
		cfg:             cfg,
		currentDate:     currentDate,
		currentNoteType: currentNoteType,
	}
}

// Resolve resolves a classified link to its actual target
func (r *Resolver) Resolve(classified ClassifiedLink) ResolvedLink {
	resolved := ResolvedLink{
		Classified:  classified,
		NeedsUpdate: false,
	}

	// Skip non-fixable links
	if !classified.NeedsFixing() {
		return resolved
	}

	switch classified.Type {
	case LinkTypeTemporalPrevious:
		return r.resolvePreviousLink(classified)
	case LinkTypeTemporalNext:
		return r.resolveNextLink(classified)
	case LinkTypeCrossReference:
		return r.resolveCrossReference(classified)
	default:
		return resolved
	}
}

// resolvePreviousLink resolves a "previous" temporal link
func (r *Resolver) resolvePreviousLink(classified ClassifiedLink) ResolvedLink {
	resolved := ResolvedLink{
		Classified: classified,
	}

	// Determine target note type
	targetType := r.determineTargetNoteType(classified)

	// Get directory for target note type
	dir, err := r.getDirForNoteType(targetType)
	if err != nil {
		resolved.Error = err
		return resolved
	}

	// Find previous note - start searching from the day before current
	searchDate := r.currentDate.AddDate(0, 0, -1)
	path, err := notes.FindNoteByDate(
		searchDate,
		targetType,
		dir,
		r.cfg.SearchWindowDays,
	)
	if err != nil {
		resolved.Error = fmt.Errorf("failed to find previous note: %w", err)
		return resolved
	}

	// Extract date from path
	date, err := notes.ParseDateFromFilename(path)
	if err != nil {
		resolved.Error = fmt.Errorf("failed to parse date from path: %w", err)
		return resolved
	}

	resolved.ResolvedPath = path
	resolved.ResolvedDate = date

	// Check if link needs updating
	currentDest := classified.Link.GetDateFromDestination()
	suggestedDest := r.formatDestination(date, targetType)

	if currentDest != date.Format(notes.DateFormat) {
		resolved.NeedsUpdate = true
		resolved.SuggestedDestination = suggestedDest
	}

	return resolved
}

// resolveNextLink resolves a "next" temporal link
func (r *Resolver) resolveNextLink(classified ClassifiedLink) ResolvedLink {
	resolved := ResolvedLink{
		Classified: classified,
	}

	// Determine target note type
	targetType := r.determineTargetNoteType(classified)

	// Get directory for target note type
	dir, err := r.getDirForNoteType(targetType)
	if err != nil {
		resolved.Error = err
		return resolved
	}

	// Find next note
	path, err := notes.FindNextNote(
		r.currentDate,
		targetType,
		dir,
		r.cfg.SearchWindowDays,
	)
	if err != nil {
		resolved.Error = fmt.Errorf("failed to find next note: %w", err)
		return resolved
	}

	// Extract date from path
	date, err := notes.ParseDateFromFilename(path)
	if err != nil {
		resolved.Error = fmt.Errorf("failed to parse date from path: %w", err)
		return resolved
	}

	resolved.ResolvedPath = path
	resolved.ResolvedDate = date

	// Check if link needs updating
	currentDest := classified.Link.GetDateFromDestination()
	suggestedDest := r.formatDestination(date, targetType)

	if currentDest != date.Format(notes.DateFormat) {
		resolved.NeedsUpdate = true
		resolved.SuggestedDestination = suggestedDest
	}

	return resolved
}

// resolveCrossReference resolves a cross-reference link (e.g., journal -> standup)
func (r *Resolver) resolveCrossReference(classified ClassifiedLink) ResolvedLink {
	resolved := ResolvedLink{
		Classified: classified,
	}

	// Determine target note type
	targetType := r.determineTargetNoteType(classified)
	if targetType == "" {
		// Can't determine target type, try to infer from current type
		if r.currentNoteType == notes.NoteTypeJournal {
			targetType = notes.NoteTypeStandup
		} else {
			targetType = notes.NoteTypeJournal
		}
	}

	// Get directory for target note type
	dir, err := r.getDirForNoteType(targetType)
	if err != nil {
		resolved.Error = err
		return resolved
	}

	// Find note for the same date (cross-reference)
	path, err := notes.FindNoteByDate(
		r.currentDate,
		targetType,
		dir,
		r.cfg.SearchWindowDays,
	)
	if err != nil {
		resolved.Error = fmt.Errorf("failed to find cross-reference note: %w", err)
		return resolved
	}

	// Extract date from path
	date, err := notes.ParseDateFromFilename(path)
	if err != nil {
		resolved.Error = fmt.Errorf("failed to parse date from path: %w", err)
		return resolved
	}

	resolved.ResolvedPath = path
	resolved.ResolvedDate = date

	// Check if link needs updating
	currentDest := classified.Link.GetDateFromDestination()
	suggestedDest := r.formatDestination(date, targetType)

	if currentDest != date.Format(notes.DateFormat) {
		resolved.NeedsUpdate = true
		resolved.SuggestedDestination = suggestedDest
	}

	return resolved
}

// determineTargetNoteType determines the target note type from the classified link
func (r *Resolver) determineTargetNoteType(classified ClassifiedLink) notes.NoteType {
	// If we have a target note type from the link destination, use it
	if classified.TargetNoteType != "" {
		return notes.NoteType(classified.TargetNoteType)
	}

	// Otherwise, for temporal links, assume same type as current note
	if classified.Type == LinkTypeTemporalPrevious || classified.Type == LinkTypeTemporalNext {
		return r.currentNoteType
	}

	// For cross-references, default to opposite type
	if r.currentNoteType == notes.NoteTypeJournal {
		return notes.NoteTypeStandup
	}
	return notes.NoteTypeJournal
}

// getDirForNoteType returns the directory path for a given note type
func (r *Resolver) getDirForNoteType(noteType notes.NoteType) (string, error) {
	switch noteType {
	case notes.NoteTypeJournal:
		return r.cfg.JournalDir()
	case notes.NoteTypeStandup:
		return r.cfg.StandupDir()
	default:
		return "", fmt.Errorf("unknown note type: %s", noteType)
	}
}

// formatDestination formats a date and note type into a link destination
// Uses relative path format: ../notetype/YYYY-MM-DD
func (r *Resolver) formatDestination(date time.Time, targetType notes.NoteType) string {
	// If target is same type as current, use simple date
	if targetType == r.currentNoteType {
		return date.Format(notes.DateFormat)
	}

	// Otherwise use relative path
	return filepath.Join("..", string(targetType), date.Format(notes.DateFormat))
}

// ResolveAll resolves all classified links
func (r *Resolver) ResolveAll(classified []ClassifiedLink) []ResolvedLink {
	resolved := make([]ResolvedLink, 0, len(classified))
	for _, c := range classified {
		resolved = append(resolved, r.Resolve(c))
	}
	return resolved
}

// FilterNeedsUpdate filters resolved links to only those that need updating
func FilterNeedsUpdate(resolved []ResolvedLink) []ResolvedLink {
	var filtered []ResolvedLink
	for _, r := range resolved {
		if r.NeedsUpdate {
			filtered = append(filtered, r)
		}
	}
	return filtered
}
