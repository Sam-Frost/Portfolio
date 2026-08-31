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

// Category splits sessions into the two buckets the daily bar chart colours
// separately. There are deliberately only these two.
type Category string

const (
	CategoryProfessional Category = "professional"
	CategoryPersonal     Category = "personal"
)

func (c Category) valid() bool {
	return c == CategoryProfessional || c == CategoryPersonal
}

// Goal is one checklist bullet: text is set when the session starts, Done
// is ticked (or not) when it ends.
type Goal struct {
	Text string `json:"text"`
	Done bool   `json:"done"`
}

// WorkSession is one "hourly tracker" timer run: a planned duration and
// category picked at start, a checklist of goals and an optional remark
// set at start, and — once it ends, one way or another — how long it
// actually ran plus the ticked-off goals and a closing remark.
type WorkSession struct {
	ID             string     `json:"id"`
	PlannedMinutes int        `json:"plannedMinutes"`
	Category       Category   `json:"category"`
	StartedAt      time.Time  `json:"startedAt"`
	EndedAt        *time.Time `json:"endedAt"`
	Status         Status     `json:"status"`
	Goals          []Goal     `json:"goals"`
	StartNote      *string    `json:"startNote"`
	Note           *string    `json:"note"`
	ActualMinutes  *int       `json:"actualMinutes"`
}

type StartInput struct {
	PlannedMinutes int      `json:"plannedMinutes"`
	Category       Category `json:"category"`
	// Goals are plain bullet strings here — all start not-done.
	Goals     []string `json:"goals"`
	StartNote string   `json:"startNote"`
}

// FinishBody is the shared shape of the complete/cancel request bodies:
// the goals with their done flags ticked and a closing remark. Complete
// requires at least one goal or a non-empty note; cancel accepts an empty
// body.
type FinishBody struct {
	Goals []Goal `json:"goals"`
	Note  string `json:"note"`
}

// DailySummary is one IST calendar day's worked minutes for the bar chart,
// split by category (the two always sum to WorkedMinutes). A session that
// crosses midnight IST contributes to both days it touched (see
// Service.DailySummary).
type DailySummary struct {
	Date                string `json:"date"`
	WorkedMinutes       int    `json:"workedMinutes"`
	ProfessionalMinutes int    `json:"professionalMinutes"`
	PersonalMinutes     int    `json:"personalMinutes"`
}
