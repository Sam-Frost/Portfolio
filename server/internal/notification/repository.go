package notification

import "context"

// Repository is the persistence boundary for push subscriptions and the
// at-most-once-per-day send ledger. Shaped for the Postgres implementation
// (repository_postgres.go); repository_memory.go is the test stand-in.
type Repository interface {
	// Subscribe upserts by endpoint — re-subscribing the same browser
	// refreshes its keys rather than creating a duplicate row.
	Subscribe(ctx context.Context, sub PushSubscription) error
	// Resync handles a rotated push subscription: if a row with oldEndpoint
	// exists it's moved to sub's new endpoint/keys, otherwise sub is
	// upserted like Subscribe. Used by the unauthenticated
	// pushsubscriptionchange path, where the browser has swapped the
	// endpoint out from under a still-valid registration.
	Resync(ctx context.Context, oldEndpoint string, sub PushSubscription) error
	// Unsubscribe removes the subscription with this endpoint; a missing
	// endpoint is not an error.
	Unsubscribe(ctx context.Context, endpoint string) error
	ListSubscriptions(ctx context.Context) ([]PushSubscription, error)

	// LogExists reports whether a notification of kind was already recorded
	// for istDate ("YYYY-MM-DD").
	LogExists(ctx context.Context, kind, istDate string) (bool, error)
	// InsertLog records that kind was sent for istDate. A duplicate
	// (kind, istDate) is swallowed so a race can't double-send.
	InsertLog(ctx context.Context, kind, istDate string) error
}
