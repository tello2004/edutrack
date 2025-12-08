package edutrack

import "time"

// DaysToYears converts a number of days to a time.Duration.
// This is useful for license durations.
func DaysToYears(days int) time.Duration {
	return time.Duration(days) * 24 * time.Hour
}

// DaysToDuration is an alias for DaysToYears for clarity.
func DaysToDuration(days int) time.Duration {
	return DaysToYears(days)
}

// YearsToDuration converts years to a time.Duration.
func YearsToDuration(years int) time.Duration {
	return DaysToYears(years * 365)
}

// MonthsToDuration converts months to a time.Duration (approximation: 30 days per month).
func MonthsToDuration(months int) time.Duration {
	return DaysToYears(months * 30)
}
