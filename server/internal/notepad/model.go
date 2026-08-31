package notepad

import "time"

// Note is the full note, content included — returned by Create, Get, and
// Update. List returns NoteSummary instead so the notepad list view isn't
// forced to pull every note's body over the wire.
type Note struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	ContentHTML string    `json:"contentHtml"`
	Pinned      bool      `json:"pinned"`
	Archived    bool      `json:"archived"`
	Locked      bool      `json:"locked"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Scratch marks the singleton "Random Notepad" jot buffer (see
	// Repository.Scratch). It's persistence-internal — it only tells List to
	// skip the note — so it's not part of the JSON contract.
	Scratch bool `json:"-"`
}

type NoteSummary struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Pinned    bool      `json:"pinned"`
	Archived  bool      `json:"archived"`
	Locked    bool      `json:"locked"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ListFilter selects which slice of notes List returns. Archived == false
// (the default) is the working set shown in the notepad; Archived == true
// is the archive view. The two are always disjoint.
type ListFilter struct {
	Archived bool
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
	Pinned      *bool   `json:"pinned"`
	Archived    *bool   `json:"archived"`
	Locked      *bool   `json:"locked"`
}

func updatedAtNow() time.Time { return time.Now().UTC() }
