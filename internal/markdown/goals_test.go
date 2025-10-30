package markdown

import (
	"testing"
)

func TestParseCheckboxItems(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []CheckboxItem
	}{
		{
			name: "mixed checkboxes",
			content: `- [ ] Pending task 1
- [x] Completed task
- [ ] Pending task 2
- [X] Another completed task`,
			expected: []CheckboxItem{
				{Checked: false, Text: "Pending task 1"},
				{Checked: true, Text: "Completed task"},
				{Checked: false, Text: "Pending task 2"},
				{Checked: true, Text: "Another completed task"},
			},
		},
		{
			name:    "all pending",
			content: "- [ ] Task 1\n- [ ] Task 2",
			expected: []CheckboxItem{
				{Checked: false, Text: "Task 1"},
				{Checked: false, Text: "Task 2"},
			},
		},
		{
			name:    "all completed",
			content: "- [x] Task 1\n- [x] Task 2",
			expected: []CheckboxItem{
				{Checked: true, Text: "Task 1"},
				{Checked: true, Text: "Task 2"},
			},
		},
		{
			name:     "no checkboxes",
			content:  "Regular bullet points\n- Item 1\n- Item 2",
			expected: []CheckboxItem{},
		},
		{
			name:     "empty content",
			content:  "",
			expected: []CheckboxItem{},
		},
		{
			name:    "with indentation",
			content: "  - [ ] Indented task\n    - [x] More indented",
			expected: []CheckboxItem{
				{Checked: false, Text: "Indented task"},
				{Checked: true, Text: "More indented"},
			},
		},
		{
			name: "malformed checkboxes without space",
			content: `- [] get pagination working
- [x] office
- [] review PR`,
			expected: []CheckboxItem{
				{Checked: false, Text: "get pagination working"},
				{Checked: true, Text: "office"},
				{Checked: false, Text: "review PR"},
			},
		},
		{
			name: "asterisk bullets with malformed checkboxes",
			content: `* [] get pagination working
* [x] office
* Check Slack messages`,
			expected: []CheckboxItem{
				{Checked: false, Text: "get pagination working"},
				{Checked: true, Text: "office"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := ParseCheckboxItems(tt.content)

			if len(items) != len(tt.expected) {
				t.Fatalf("expected %d items, got %d", len(tt.expected), len(items))
			}

			for i, item := range items {
				if item.Checked != tt.expected[i].Checked {
					t.Errorf("item %d: expected Checked=%v, got %v", i, tt.expected[i].Checked, item.Checked)
				}
				if item.Text != tt.expected[i].Text {
					t.Errorf("item %d: expected Text=%q, got %q", i, tt.expected[i].Text, item.Text)
				}
			}
		})
	}
}

func TestFilterPendingItems(t *testing.T) {
	items := []CheckboxItem{
		{Checked: false, Text: "Pending 1"},
		{Checked: true, Text: "Done"},
		{Checked: false, Text: "Pending 2"},
		{Checked: true, Text: "Also done"},
	}

	pending := FilterPendingItems(items)

	if len(pending) != 2 {
		t.Fatalf("expected 2 pending items, got %d", len(pending))
	}

	if pending[0].Text != "Pending 1" || pending[1].Text != "Pending 2" {
		t.Errorf("wrong pending items: got %v", pending)
	}
}

func TestFormatCheckboxItems(t *testing.T) {
	tests := []struct {
		name     string
		items    []CheckboxItem
		expected string
	}{
		{
			name: "mixed items",
			items: []CheckboxItem{
				{Checked: false, Text: "Pending task"},
				{Checked: true, Text: "Done task"},
			},
			expected: "- [ ] Pending task\n- [x] Done task",
		},
		{
			name:     "empty items",
			items:    []CheckboxItem{},
			expected: "",
		},
		{
			name: "only pending",
			items: []CheckboxItem{
				{Checked: false, Text: "Task 1"},
				{Checked: false, Text: "Task 2"},
			},
			expected: "- [ ] Task 1\n- [ ] Task 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCheckboxItems(tt.items)
			if result != tt.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestParseGoalItems(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []GoalItem
	}{
		{
			name: "mixed checkboxes and plain bullets",
			content: `- [ ] Unchecked task
- [x] Completed task
- Plain bullet point
- Another plain item`,
			expected: []GoalItem{
				{Text: "Unchecked task", HasCheckbox: true, Checked: false},
				{Text: "Completed task", HasCheckbox: true, Checked: true},
				{Text: "Plain bullet point", HasCheckbox: false, Checked: false},
				{Text: "Another plain item", HasCheckbox: false, Checked: false},
			},
		},
		{
			name:    "only checkboxes",
			content: "- [ ] Task 1\n- [x] Task 2",
			expected: []GoalItem{
				{Text: "Task 1", HasCheckbox: true, Checked: false},
				{Text: "Task 2", HasCheckbox: true, Checked: true},
			},
		},
		{
			name:    "only plain bullets",
			content: "- Item 1\n- Item 2",
			expected: []GoalItem{
				{Text: "Item 1", HasCheckbox: false, Checked: false},
				{Text: "Item 2", HasCheckbox: false, Checked: false},
			},
		},
		{
			name:     "empty content",
			content:  "",
			expected: []GoalItem{},
		},
		{
			name:    "with indentation",
			content: "  - [ ] Indented checkbox\n    - Plain indented",
			expected: []GoalItem{
				{Text: "Indented checkbox", HasCheckbox: true, Checked: false},
				{Text: "Plain indented", HasCheckbox: false, Checked: false},
			},
		},
		{
			name: "malformed checkboxes without space",
			content: `- [] get pagination working
- [x] office
- Check Slack messages (plain bullet - state unknown)
- [] review PR`,
			expected: []GoalItem{
				{Text: "get pagination working", HasCheckbox: true, Checked: false},
				{Text: "office", HasCheckbox: true, Checked: true},
				{Text: "Check Slack messages (plain bullet - state unknown)", HasCheckbox: false, Checked: false},
				{Text: "review PR", HasCheckbox: true, Checked: false},
			},
		},
		{
			name: "asterisk bullets with mixed content",
			content: `* [] get pagination working
* [x] office
* Check Slack messages (plain bullet - state unknown)
* [ ] review PR`,
			expected: []GoalItem{
				{Text: "get pagination working", HasCheckbox: true, Checked: false},
				{Text: "office", HasCheckbox: true, Checked: true},
				{Text: "Check Slack messages (plain bullet - state unknown)", HasCheckbox: false, Checked: false},
				{Text: "review PR", HasCheckbox: true, Checked: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := ParseGoalItems(tt.content)

			if len(items) != len(tt.expected) {
				t.Fatalf("expected %d items, got %d", len(tt.expected), len(items))
			}

			for i, item := range items {
				if item.Text != tt.expected[i].Text {
					t.Errorf("item %d: expected Text=%q, got %q", i, tt.expected[i].Text, item.Text)
				}
				if item.HasCheckbox != tt.expected[i].HasCheckbox {
					t.Errorf("item %d: expected HasCheckbox=%v, got %v", i, tt.expected[i].HasCheckbox, item.HasCheckbox)
				}
				if item.Checked != tt.expected[i].Checked {
					t.Errorf("item %d: expected Checked=%v, got %v", i, tt.expected[i].Checked, item.Checked)
				}
			}
		})
	}
}

func TestFilterUnfinishedGoals(t *testing.T) {
	items := []GoalItem{
		{Text: "Unchecked", HasCheckbox: true, Checked: false},
		{Text: "Completed", HasCheckbox: true, Checked: true},
		{Text: "Plain bullet", HasCheckbox: false, Checked: false},
		{Text: "Another completed", HasCheckbox: true, Checked: true},
		{Text: "Another plain", HasCheckbox: false, Checked: false},
	}

	unfinished := FilterUnfinishedGoals(items)

	// Should include: unchecked (1) + plain bullets (2) = 3 items
	if len(unfinished) != 3 {
		t.Fatalf("expected 3 unfinished items, got %d", len(unfinished))
	}

	// Verify we have the right items
	expectedTexts := []string{"Unchecked", "Plain bullet", "Another plain"}
	for i, item := range unfinished {
		if item.Text != expectedTexts[i] {
			t.Errorf("item %d: expected %q, got %q", i, expectedTexts[i], item.Text)
		}
	}
}

func TestFormatGoalItems(t *testing.T) {
	tests := []struct {
		name     string
		items    []GoalItem
		expected string
	}{
		{
			name: "mixed items",
			items: []GoalItem{
				{Text: "Unchecked task", HasCheckbox: true, Checked: false},
				{Text: "Completed task", HasCheckbox: true, Checked: true},
				{Text: "Plain item", HasCheckbox: false, Checked: false},
			},
			expected: "- [ ] Unchecked task\n- [x] Completed task\n- Plain item",
		},
		{
			name:     "empty items",
			items:    []GoalItem{},
			expected: "",
		},
		{
			name: "only plain bullets",
			items: []GoalItem{
				{Text: "Item 1", HasCheckbox: false},
				{Text: "Item 2", HasCheckbox: false},
			},
			expected: "- Item 1\n- Item 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatGoalItems(tt.items)
			if result != tt.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.expected, result)
			}
		})
	}
}
