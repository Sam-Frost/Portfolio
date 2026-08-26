package diary

import "time"

// EntryDateLayout is the wire and DB-query format for an Entry's calendar
// date ("YYYY-MM-DD"), matching todo.TargetDateLayout's convention. It's
// also the layout entry_date is stored under (a DATE column, not a
// timestamp) since a diary entry's identity is a calendar day, not an
// instant.
const EntryDateLayout = "2006-01-02"

// IST is the fixed timezone a diary entry's calendar day (and its edit-lock
// deadline) is measured in, resolved once at package init rather than
// trusting the server process's own local timezone/wall clock — see
// IsLocked. A failure to resolve it panics at startup instead of silently
// producing wrong lock decisions on every request.
var IST = mustLoadIST()

func mustLoadIST() *time.Location {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		panic("diary: failed to load Asia/Kolkata timezone: " + err.Error())
	}
	return loc
}

// Entry is one calendar day's diary entry. Content is HTML — the diary
// reuses the notepad feature's rich-text editor, so it produces the same
// contentEditable innerHTML notepad's Note.ContentHTML stores.
//
// Locked is not persisted; it's computed fresh on every read/write by the
// service (see IsLocked) from EntryDate and the current instant, so it's
// always accurate even for an entry that was fetched a while ago.
type Entry struct {
	ID        string    `json:"id"`
	EntryDate string    `json:"entryDate"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Locked    bool      `json:"locked"`
}

// UpsertInput is the body of a PUT — an entry for a date that already
// exists is updated in place ("one entry per day", edited over time) rather
// than rejected as a duplicate.
type UpsertInput struct {
	Content string `json:"content"`
}

// IsLocked reports whether entryDate's edit window has closed as of now.
// The rule: an entry becomes uneditable once the calendar day it belongs to
// (IST) has been over for 24 hours. A day is "over" at IST midnight
// starting the next day, so the lock instant is IST midnight two days after
// entryDate — e.g. 2026-08-20 locks at 2026-08-22T00:00:00+05:30.
//
// now is compared as an absolute instant (time.Time comparisons are
// timezone-independent), so it's fine to pass e.g. time.Now().UTC().
func IsLocked(entryDate string, now time.Time) bool {
	d, err := time.ParseInLocation(EntryDateLayout, entryDate, IST)
	if err != nil {
		// A malformed date shouldn't be reachable (service validates first),
		// but treat it as locked rather than editable if it ever is.
		return true
	}
	lockAt := d.AddDate(0, 0, 2)
	return !now.Before(lockAt)
}
