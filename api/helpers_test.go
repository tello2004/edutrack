package edutrack

import (
	"testing"
	"time"
)

func TestDaysToYears(t *testing.T) {
	tests := []struct {
		name string
		days int
		want time.Duration
	}{
		{
			name: "1 day",
			days: 1,
			want: 24 * time.Hour,
		},
		{
			name: "7 days (week)",
			days: 7,
			want: 7 * 24 * time.Hour,
		},
		{
			name: "30 days (month)",
			days: 30,
			want: 30 * 24 * time.Hour,
		},
		{
			name: "365 days (year)",
			days: 365,
			want: 365 * 24 * time.Hour,
		},
		{
			name: "0 days",
			days: 0,
			want: 0,
		},
		{
			name: "negative days",
			days: -10,
			want: -10 * 24 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DaysToYears(tt.days); got != tt.want {
				t.Errorf("DaysToYears(%d) = %v, want %v", tt.days, got, tt.want)
			}
		})
	}
}

func TestDaysToDuration(t *testing.T) {
	// DaysToDuration should be an alias for DaysToYears
	tests := []struct {
		name string
		days int
		want time.Duration
	}{
		{
			name: "1 day",
			days: 1,
			want: 24 * time.Hour,
		},
		{
			name: "30 days",
			days: 30,
			want: 30 * 24 * time.Hour,
		},
		{
			name: "365 days",
			days: 365,
			want: 365 * 24 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DaysToDuration(tt.days)
			if got != tt.want {
				t.Errorf("DaysToDuration(%d) = %v, want %v", tt.days, got, tt.want)
			}

			// Ensure it matches DaysToYears
			if got != DaysToYears(tt.days) {
				t.Errorf("DaysToDuration(%d) != DaysToYears(%d)", tt.days, tt.days)
			}
		})
	}
}

func TestYearsToDuration(t *testing.T) {
	tests := []struct {
		name  string
		years int
		want  time.Duration
	}{
		{
			name:  "1 year",
			years: 1,
			want:  365 * 24 * time.Hour,
		},
		{
			name:  "2 years",
			years: 2,
			want:  2 * 365 * 24 * time.Hour,
		},
		{
			name:  "5 years",
			years: 5,
			want:  5 * 365 * 24 * time.Hour,
		},
		{
			name:  "0 years",
			years: 0,
			want:  0,
		},
		{
			name:  "negative years",
			years: -1,
			want:  -365 * 24 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := YearsToDuration(tt.years); got != tt.want {
				t.Errorf("YearsToDuration(%d) = %v, want %v", tt.years, got, tt.want)
			}
		})
	}
}

func TestMonthsToDuration(t *testing.T) {
	tests := []struct {
		name   string
		months int
		want   time.Duration
	}{
		{
			name:   "1 month (30 days)",
			months: 1,
			want:   30 * 24 * time.Hour,
		},
		{
			name:   "6 months",
			months: 6,
			want:   6 * 30 * 24 * time.Hour,
		},
		{
			name:   "12 months",
			months: 12,
			want:   12 * 30 * 24 * time.Hour,
		},
		{
			name:   "0 months",
			months: 0,
			want:   0,
		},
		{
			name:   "negative months",
			months: -3,
			want:   -3 * 30 * 24 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MonthsToDuration(tt.months); got != tt.want {
				t.Errorf("MonthsToDuration(%d) = %v, want %v", tt.months, got, tt.want)
			}
		})
	}
}

func TestDurationConversionsConsistency(t *testing.T) {
	// 12 months should equal 360 days (30 * 12)
	months12 := MonthsToDuration(12)
	days360 := DaysToDuration(360)
	if months12 != days360 {
		t.Errorf("MonthsToDuration(12) = %v, DaysToDuration(360) = %v, want equal", months12, days360)
	}

	// 1 year should equal 365 days
	year1 := YearsToDuration(1)
	days365 := DaysToDuration(365)
	if year1 != days365 {
		t.Errorf("YearsToDuration(1) = %v, DaysToDuration(365) = %v, want equal", year1, days365)
	}
}

func TestDurationConversionsWithTimeNow(t *testing.T) {
	now := time.Now()

	// Adding 30 days using DaysToDuration
	future := now.Add(DaysToDuration(30))
	expected := now.Add(30 * 24 * time.Hour)

	if !future.Equal(expected) {
		t.Errorf("time.Now().Add(DaysToDuration(30)) = %v, want %v", future, expected)
	}

	// Adding 1 year using YearsToDuration
	futureYear := now.Add(YearsToDuration(1))
	expectedYear := now.Add(365 * 24 * time.Hour)

	if !futureYear.Equal(expectedYear) {
		t.Errorf("time.Now().Add(YearsToDuration(1)) = %v, want %v", futureYear, expectedYear)
	}
}

func BenchmarkDaysToYears(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DaysToYears(365)
	}
}

func BenchmarkYearsToDuration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = YearsToDuration(5)
	}
}

func BenchmarkMonthsToDuration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = MonthsToDuration(12)
	}
}
