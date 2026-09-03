package reminder

import "time"

// Kind is how a reminder repeats.
type Kind string

const (
	// KindOnce fires at FireAt, then is deleted.
	KindOnce Kind = "once"
	// KindRepeat fires every IntervalSeconds; FireAt holds the next
	// occurrence and is advanced after each fire. It keeps firing until the
	// reminder is deleted or its todo is marked done (the scheduler's due
	// query joins todos and skips done ones).
	KindRepeat Kind = "repeat"
)

// MinIntervalSeconds is the floor for a repeating reminder — the scheduler
// ticks once a minute, so anything shorter can't be honoured anyway.
const MinIntervalSeconds = 60

// Reminder is one alarm attached to a todo.
type Reminder struct {
	ID              string    `json:"id"`
	TodoID          string    `json:"todoId"`
	Kind            Kind      `json:"kind"`
	FireAt          time.Time `json:"fireAt"`
	IntervalSeconds *int      `json:"intervalSeconds"` // nil for KindOnce
	CreatedAt       time.Time `json:"createdAt"`
}

// CreateInput is the client payload. For KindOnce the client sends an
// absolute FireAt (it computes "+15m" / "+2 days" itself). For KindRepeat
// it sends only IntervalSeconds and the server sets the first FireAt to
// now + interval.
type CreateInput struct {
	Kind            Kind    `json:"kind"`
	FireAt          *string `json:"fireAt"` // RFC3339; required for "once"
	IntervalSeconds *int    `json:"intervalSeconds"`
}

// Due is a reminder that's ready to fire, carrying its todo's name for the
// notification text.
type Due struct {
	Reminder
	TodoName string
}
