package notification

import (
	"context"
	"testing"

	"github.com/Sam-Frost/portfolio/internal/mailer"
	"github.com/Sam-Frost/portfolio/internal/settings"
)

type staticSettings struct{}

func (staticSettings) Get(context.Context) (settings.Settings, error) {
	return settings.Settings{}, nil
}

func newTestService(repo Repository) *Service {
	return NewService(repo, NewNoopPushSender(), mailer.NewNoopMailer(), staticSettings{})
}

func sub(endpoint string) SubscribeInput {
	return SubscribeInput{Endpoint: endpoint, P256dh: "p", Auth: "a"}
}

func endpoints(t *testing.T, repo *MemoryRepository) []string {
	t.Helper()
	list, err := repo.ListSubscriptions(context.Background())
	if err != nil {
		t.Fatalf("ListSubscriptions: %v", err)
	}
	out := make([]string, len(list))
	for i, s := range list {
		out[i] = s.Endpoint
	}
	return out
}

func TestResync_MovesRotatedSubscription(t *testing.T) {
	repo := NewMemoryRepository()
	svc := newTestService(repo)
	ctx := context.Background()

	if err := svc.Subscribe(ctx, sub("https://push/old")); err != nil {
		t.Fatal(err)
	}
	if err := svc.Resync(ctx, "https://push/old", sub("https://push/new")); err != nil {
		t.Fatalf("Resync: %v", err)
	}

	got := endpoints(t, repo)
	if len(got) != 1 || got[0] != "https://push/new" {
		t.Fatalf("endpoints = %v, want just the rotated one", got)
	}
}

func TestResync_UnknownOldEndpointUpserts(t *testing.T) {
	repo := NewMemoryRepository()
	svc := newTestService(repo)
	ctx := context.Background()

	if err := svc.Resync(ctx, "https://push/gone", sub("https://push/new")); err != nil {
		t.Fatalf("Resync: %v", err)
	}
	got := endpoints(t, repo)
	if len(got) != 1 || got[0] != "https://push/new" {
		t.Fatalf("endpoints = %v, want the new one inserted", got)
	}
}

func TestResync_RejectsIncompletePayload(t *testing.T) {
	svc := newTestService(NewMemoryRepository())
	if err := svc.Resync(context.Background(), "", SubscribeInput{Endpoint: "https://push/x"}); err == nil {
		t.Fatal("expected an error for a payload missing keys")
	}
}
