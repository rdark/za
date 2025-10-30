package util

import "time"

// IsSameWeek returns true if both dates are in the same week (Monday-Sunday)
func IsSameWeek(date1, date2 time.Time) bool {
	// Get the ISO week (year, week number)
	year1, week1 := date1.ISOWeek()
	year2, week2 := date2.ISOWeek()

	return year1 == year2 && week1 == week2
}

// IsWeekday returns true if the date is a weekday (Monday-Friday)
func IsWeekday(date time.Time) bool {
	weekday := date.Weekday()
	return weekday >= time.Monday && weekday <= time.Friday
}
