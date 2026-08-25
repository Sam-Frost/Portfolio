package settings

type DailyWorkTracker struct {
	TotalWorkHoursRequired *float64 `json:"totalWorkHoursRequired"`
}

// TimeLeftFormat controls how the countdown to TimeLeftClock.GoalDate is
// broken down for display.
type TimeLeftFormat string

const (
	FormatWeeksDaysTime TimeLeftFormat = "weeks_days_time"
	FormatDaysTime      TimeLeftFormat = "days_time"
)

// TimeLeftClock is the single configured countdown target shown across the
// domain area (login popup + persistent top bar clock). GoalDate is an
// RFC3339 timestamp string, nil when unset.
type TimeLeftClock struct {
	GoalDate *string        `json:"goalDate"`
	Format   TimeLeftFormat `json:"format"`
}

type Settings struct {
	DailyWorkTracker DailyWorkTracker `json:"dailyWorkTracker"`
	TimeLeftClock    TimeLeftClock    `json:"timeLeftClock"`
}

// DailyWorkTrackerInput mirrors DailyWorkTracker: the frontend always sends
// the whole nested object, so there's no tri-state (missing vs. explicit
// null) to track beyond the pointer already needed for "no value set".
type DailyWorkTrackerInput struct {
	TotalWorkHoursRequired *float64 `json:"totalWorkHoursRequired"`
}

// TimeLeftClockInput mirrors TimeLeftClock for the same reason.
type TimeLeftClockInput struct {
	GoalDate *string        `json:"goalDate"`
	Format   TimeLeftFormat `json:"format"`
}

// UpdateInput is a partial update: a nil section leaves that section
// unchanged, matching the label package's UpdateInput convention.
type UpdateInput struct {
	DailyWorkTracker *DailyWorkTrackerInput `json:"dailyWorkTracker"`
	TimeLeftClock    *TimeLeftClockInput    `json:"timeLeftClock"`
}
