package notepad

import "time"

// Note is the full note, content included — returned by Create, Get, and
// Update. List returns NoteSummary instead so the notepad list view isn't
// forced to pull every note's body over the wire.
type Note struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	ContentHTML string    `json:"contentHtml"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type NoteSummary struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// DefaultTitleLayout is used to title a note when it's created without one:
// the note's created-at timestamp, e.g. "Jan 2, 2006 3:04 PM".
const DefaultTitleLayout = "Jan 2, 2006 3:04 PM"

type CreateInput struct {
	Title *string `json:"title"`
}

// UpdateInput is a partial update: nil fields are left unchanged, matching
// the todo/label packages' UpdateInput convention. Autosave sends whichever
// of Title/ContentHTML actually changed.
type UpdateInput struct {
	Title       *string `json:"title"`
	ContentHTML *string `json:"contentHtml"`
}

func updatedAtNow() time.Time { return time.Now().UTC() }
