package diary

import "context"

// Repository is the persistence boundary for diary entries, shaped for the
// eventual Postgres implementation (see repository_postgres.go);
// repository_memory.go is the in-memory stand-in used by tests. Upsert
// takes the date and content directly (not a mutation closure) so a SQL
// implementation can express "one entry per day" as a single
// INSERT ... ON CONFLICT (entry_date) DO UPDATE rather than a
// read-then-write.
//
// Lock enforcement is deliberately not the repository's job — it has no
// notion of "now" or timezone; the service checks IsLocked before ever
// calling Upsert.
type Repository interface {
	GetByDate(ctx context.Context, date string) (Entry, error)
	Upsert(ctx context.Context, date string, content string) (Entry, error)
	// ListDates returns the entry_date (as EntryDateLayout strings) of every
	// entry in [from, to], inclusive — just enough for a calendar view to
	// mark which days have an entry, without pulling every day's content.
	ListDates(ctx context.Context, from, to string) ([]string, error)
}
