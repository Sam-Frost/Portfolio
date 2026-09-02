package drawingboard

import (
	"context"
	"testing"
	"time"
)

func TestMemory_ListOrdersByUpdatedAtDescending(t *testing.T) {
	r := NewMemoryRepository()
	ctx := context.Background()

	older, err := r.Create(ctx, Board{Name: "Older", SceneData: []byte(DefaultSceneData)})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	newer, err := r.Create(ctx, Board{Name: "Newer", SceneData: []byte(DefaultSceneData)})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Touch "older" so it becomes the most recently updated board.
	name := "Older"
	if _, err := r.Update(ctx, older.ID, UpdateInput{Name: &name}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	list, err := r.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 || list[0].ID != older.ID || list[1].ID != newer.ID {
		t.Fatalf("List order = %+v, want [older, newer]", list)
	}
}

func TestMemory_UpdateUnknownBoardReturnsNotFound(t *testing.T) {
	r := NewMemoryRepository()

	_, err := r.Update(context.Background(), "missing", UpdateInput{Name: ptr("x")})
	assertNotFound(t, err)
}

func TestMemory_DeleteUnknownBoardReturnsNotFound(t *testing.T) {
	r := NewMemoryRepository()

	err := r.Delete(context.Background(), "missing")
	assertNotFound(t, err)
}

func TestMemory_UpdateBumpsUpdatedAt(t *testing.T) {
	r := NewMemoryRepository()
	ctx := context.Background()

	b, err := r.Create(ctx, Board{Name: "Board", SceneData: []byte(DefaultSceneData)})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	before := b.UpdatedAt

	time.Sleep(time.Millisecond)
	updated, err := r.Update(ctx, b.ID, UpdateInput{Name: ptr("Renamed")})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !updated.UpdatedAt.After(before) {
		t.Errorf("UpdatedAt = %v, want after %v", updated.UpdatedAt, before)
	}
}
