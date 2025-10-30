package util

import (
	"testing"
	"time"
)

func TestIsSameWeek(t *testing.T) {
	tests := []struct {
		name     string
		date1    string
		date2    string
		expected bool
	}{
		{
			name:     "same day",
			date1:    "2025-01-13",
			date2:    "2025-01-13",
			expected: true,
		},
		{
			name:     "monday and friday same week",
			date1:    "2025-01-13", // Monday
			date2:    "2025-01-17", // Friday
			expected: true,
		},
		{
			name:     "monday and sunday same week",
			date1:    "2025-01-13", // Monday
			date2:    "2025-01-19", // Sunday
			expected: true,
		},
		{
			name:     "friday and next monday different weeks",
			date1:    "2025-01-17", // Friday
			date2:    "2025-01-20", // Monday
			expected: false,
		},
		{
			name:     "same weekday different weeks",
			date1:    "2025-01-13", // Monday
			date2:    "2025-01-20", // Monday next week
			expected: false,
		},
		{
			name:     "end of year week boundary",
			date1:    "2024-12-30", // Monday, week 1 of 2025
			date2:    "2025-01-05", // Sunday, week 1 of 2025
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d1, _ := time.Parse("2006-01-02", tt.date1)
			d2, _ := time.Parse("2006-01-02", tt.date2)

			result := IsSameWeek(d1, d2)
			if result != tt.expected {
				t.Errorf("IsSameWeek(%s, %s) = %v, expected %v",
					tt.date1, tt.date2, result, tt.expected)
			}
		})
	}
}

func TestIsWeekday(t *testing.T) {
	tests := []struct {
		name     string
		date     string
		expected bool
	}{
		{
			name:     "Monday is weekday",
			date:     "2025-02-03", // Monday
			expected: true,
		},
		{
			name:     "Tuesday is weekday",
			date:     "2025-02-04", // Tuesday
			expected: true,
		},
		{
			name:     "Wednesday is weekday",
			date:     "2025-02-05", // Wednesday
			expected: true,
		},
		{
			name:     "Thursday is weekday",
			date:     "2025-02-06", // Thursday
			expected: true,
		},
		{
			name:     "Friday is weekday",
			date:     "2025-02-07", // Friday
			expected: true,
		},
		{
			name:     "Saturday is not weekday",
			date:     "2025-02-08", // Saturday
			expected: false,
		},
		{
			name:     "Sunday is not weekday",
			date:     "2025-02-09", // Sunday
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, _ := time.Parse("2006-01-02", tt.date)
			result := IsWeekday(d)
			if result != tt.expected {
				t.Errorf("IsWeekday(%s [%v]) = %v, expected %v",
					tt.date, d.Weekday(), result, tt.expected)
			}
		})
	}
}
