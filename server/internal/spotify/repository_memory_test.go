package spotify

import (
	"context"
	"testing"
	"time"
)

func TestMemoryRepository_GetNotConnectedInitially(t *testing.T) {
	repo := NewMemoryRepository()

	_, ok, err := repo.Get(context.Background())
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if ok {
		t.Error("ok = true, want false before any Save")
	}
}

func TestMemoryRepository_SaveThenGetRoundTrips(t *testing.T) {
	repo := NewMemoryRepository()
	want := TokenSet{
		RefreshToken:         "refresh-abc",
		AccessToken:          "access-xyz",
		AccessTokenExpiresAt: time.Now().Add(time.Hour).Truncate(time.Second),
	}

	if err := repo.Save(context.Background(), want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, ok, err := repo.Get(context.Background())
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok {
		t.Fatal("ok = false, want true after Save")
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestMemoryRepository_SaveOverwritesPreviousTokens(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	_ = repo.Save(ctx, TokenSet{RefreshToken: "first", AccessToken: "a1"})
	_ = repo.Save(ctx, TokenSet{RefreshToken: "second", AccessToken: "a2"})

	got, _, _ := repo.Get(ctx)
	if got.RefreshToken != "second" {
		t.Errorf("RefreshToken = %q, want %q (most recent Save)", got.RefreshToken, "second")
	}
}
