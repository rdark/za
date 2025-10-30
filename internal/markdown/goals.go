package markdown

import (
	"regexp"
	"strings"
)

var (
	// Regex to match checkbox items: - [ ], * [ ], - [], * [], - [x], * [x], etc.
	// Handles both - and * bullet markers, and both well-formed [ ] and malformed [] (no space)
	checkboxRegex = regexp.MustCompile(`^\s*[-*]\s*\[([\ xX]*)\]\s*(.+)$`)
	// Regex to match plain bullet points: - item or * item
	bulletRegex = regexp.MustCompile(`^\s*[-*]\s+(.+)$`)
)

// CheckboxItem represents a task with a checkbox
type CheckboxItem struct {
	Checked bool
	Text    string
}

// GoalItem represents a goal that can be either a checkbox item or plain bullet point
type GoalItem struct {
	Text        string
	HasCheckbox bool
	Checked     bool // Only meaningful if HasCheckbox is true
}

// ParseCheckboxItems extracts checkbox items from content
func ParseCheckboxItems(content string) []CheckboxItem {
	var items []CheckboxItem

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if matches := checkboxRegex.FindStringSubmatch(line); matches != nil {
			// Check if the checkbox contains 'x' or 'X' to determine if checked
			checkboxContent := strings.ToLower(strings.TrimSpace(matches[1]))
			checked := strings.Contains(checkboxContent, "x")
			text := strings.TrimSpace(matches[2])
			items = append(items, CheckboxItem{
				Checked: checked,
				Text:    text,
			})
		}
	}

	return items
}

// FilterPendingItems returns only unchecked items
func FilterPendingItems(items []CheckboxItem) []CheckboxItem {
	var pending []CheckboxItem
	for _, item := range items {
		if !item.Checked {
			pending = append(pending, item)
		}
	}
	return pending
}

// FormatCheckboxItems converts items back to markdown checkbox format
func FormatCheckboxItems(items []CheckboxItem) string {
	if len(items) == 0 {
		return ""
	}

	var lines []string
	for _, item := range items {
		checkbox := "[ ]"
		if item.Checked {
			checkbox = "[x]"
		}
		lines = append(lines, "- "+checkbox+" "+item.Text)
	}

	return strings.Join(lines, "\n")
}

// ParseGoalItems extracts both checkbox items and plain bullet points from content
func ParseGoalItems(content string) []GoalItem {
	var items []GoalItem

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		// First try to match checkbox items
		if matches := checkboxRegex.FindStringSubmatch(line); matches != nil {
			// Check if the checkbox contains 'x' or 'X' to determine if checked
			checkboxContent := strings.ToLower(strings.TrimSpace(matches[1]))
			checked := strings.Contains(checkboxContent, "x")
			text := strings.TrimSpace(matches[2])
			items = append(items, GoalItem{
				Text:        text,
				HasCheckbox: true,
				Checked:     checked,
			})
			continue
		}

		// Then try to match plain bullet points
		if matches := bulletRegex.FindStringSubmatch(line); matches != nil {
			text := strings.TrimSpace(matches[1])
			// Skip if it looks like a checkbox we missed (shouldn't happen)
			if strings.HasPrefix(text, "[") {
				continue
			}
			items = append(items, GoalItem{
				Text:        text,
				HasCheckbox: false,
				Checked:     false,
			})
		}
	}

	return items
}

// FilterUnfinishedGoals returns items that should be copied forward:
// - Unchecked checkbox items [ ]
// - Plain bullet points without checkboxes (unknown state)
// Does NOT include checked items [x]
func FilterUnfinishedGoals(items []GoalItem) []GoalItem {
	var unfinished []GoalItem
	for _, item := range items {
		// Include if it's not a checkbox (plain bullet)
		// OR if it's a checkbox that's not checked
		if !item.HasCheckbox || !item.Checked {
			unfinished = append(unfinished, item)
		}
	}
	return unfinished
}

// FormatGoalItems converts goal items back to markdown format
func FormatGoalItems(items []GoalItem) string {
	if len(items) == 0 {
		return ""
	}

	var lines []string
	for _, item := range items {
		if item.HasCheckbox {
			checkbox := "[ ]"
			if item.Checked {
				checkbox = "[x]"
			}
			lines = append(lines, "- "+checkbox+" "+item.Text)
		} else {
			lines = append(lines, "- "+item.Text)
		}
	}

	return strings.Join(lines, "\n")
}
