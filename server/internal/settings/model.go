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

// Notifications is the config for the outbound notification channels
// (email + Web Push). RecipientEmail is nil until set in the UI. MorningTime
// is the local (IST) "HH:MM" the daily overdue-todo digest is sent at. The
// two enable flags gate each delivery channel independently.
type Notifications struct {
	RecipientEmail *string `json:"recipientEmail"`
	MorningTime    string  `json:"morningTime"`
	EmailEnabled   bool    `json:"emailEnabled"`
	PushEnabled    bool    `json:"pushEnabled"`
}

type Settings struct {
	DailyWorkTracker DailyWorkTracker `json:"dailyWorkTracker"`
	TimeLeftClock    TimeLeftClock    `json:"timeLeftClock"`
	Notifications    Notifications    `json:"notifications"`
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

// NotificationsInput mirrors Notifications — the frontend sends the whole
// section, so a nil RecipientEmail here means "clear it" rather than "leave
// unchanged" (the section pointer on UpdateInput is what carries the
// leave-unchanged tri-state).
type NotificationsInput struct {
	RecipientEmail *string `json:"recipientEmail"`
	MorningTime    string  `json:"morningTime"`
	EmailEnabled   bool    `json:"emailEnabled"`
	PushEnabled    bool    `json:"pushEnabled"`
}

// UpdateInput is a partial update: a nil section leaves that section
// unchanged, matching the label package's UpdateInput convention.
type UpdateInput struct {
	DailyWorkTracker *DailyWorkTrackerInput `json:"dailyWorkTracker"`
	TimeLeftClock    *TimeLeftClockInput    `json:"timeLeftClock"`
	Notifications    *NotificationsInput    `json:"notifications"`
}
