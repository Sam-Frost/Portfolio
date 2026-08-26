package worksession

import "time"

// DateLayout is the wire format for calendar-day boundaries ("YYYY-MM-DD",
// interpreted as an IST calendar date — see ParseISTDayStart) used by the
// range/daily-summary query params and DailySummary.Date.
const DateLayout = "2006-01-02"

type Status string

const (
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusCancelled Status = "cancelled"
)

// WorkSession is one "hourly tracker" timer run: a planned duration picked
// at start, and — once it ends, one way or another — how long it actually
// ran and what the user says they did with it.
type WorkSession struct {
	ID             string     `json:"id"`
	PlannedMinutes int        `json:"plannedMinutes"`
	StartedAt      time.Time  `json:"startedAt"`
	EndedAt        *time.Time `json:"endedAt"`
	Status         Status     `json:"status"`
	Note           *string    `json:"note"`
	ActualMinutes  *int       `json:"actualMinutes"`
}

type StartInput struct {
	PlannedMinutes int `json:"plannedMinutes"`
}

// CompleteInput requires a note — a session that runs its full planned
// duration must be logged with what it was spent on.
type CompleteInput struct {
	Note string `json:"note"`
}

// CancelInput's note is optional — a session cut short may have nothing
// worth logging.
type CancelInput struct {
	Note *string `json:"note"`
}

// DailySummary is one IST calendar day's worked minutes, for the hourly
// tracker's bar chart. A session that crosses midnight IST contributes to
// both days it touched (see Service.DailySummary).
type DailySummary struct {
	Date          string `json:"date"`
	WorkedMinutes int    `json:"workedMinutes"`
}
