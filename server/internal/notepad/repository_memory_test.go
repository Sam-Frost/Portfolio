package notepad

import (
	"context"
	"testing"
)

// The Notepad label lives on the note row (notes.label_id). Assigning,
// reassigning, and clearing it all go through Update; per UpdateInput's
// contract an empty-string LabelID means "clear it" and a nil LabelID means
// "leave unchanged".

func TestMemory_UpdateAssignsLabel(t *testing.T) {
	r := NewMemoryRepository()
	n, err := r.Create(context.Background(), Note{Title: "Note"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	labelID := "label-123"
	updated, err := r.Update(context.Background(), n.ID, UpdateInput{LabelID: &labelID})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.LabelID == nil || *updated.LabelID != "label-123" {
		t.Fatalf("LabelID = %v, want %q", updated.LabelID, "label-123")
	}
}

func TestMemory_UpdateClearsLabelWithEmptyString(t *testing.T) {
	r := NewMemoryRepository()
	n, err := r.Create(context.Background(), Note{Title: "Note"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	labelID := "label-123"
	if _, err := r.Update(context.Background(), n.ID, UpdateInput{LabelID: &labelID}); err != nil {
		t.Fatalf("Update (assign): %v", err)
	}

	empty := ""
	cleared, err := r.Update(context.Background(), n.ID, UpdateInput{LabelID: &empty})
	if err != nil {
		t.Fatalf("Update (clear): %v", err)
	}
	if cleared.LabelID != nil {
		t.Fatalf("LabelID = %v, want nil after clear", *cleared.LabelID)
	}
}

func TestMemory_UpdateLeavesLabelUnchangedWhenNil(t *testing.T) {
	r := NewMemoryRepository()
	n, err := r.Create(context.Background(), Note{Title: "Note"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	labelID := "label-123"
	if _, err := r.Update(context.Background(), n.ID, UpdateInput{LabelID: &labelID}); err != nil {
		t.Fatalf("Update (assign): %v", err)
	}

	// An unrelated field update must not disturb the label.
	title := "Renamed"
	updated, err := r.Update(context.Background(), n.ID, UpdateInput{Title: &title})
	if err != nil {
		t.Fatalf("Update (rename): %v", err)
	}
	if updated.LabelID == nil || *updated.LabelID != "label-123" {
		t.Fatalf("LabelID = %v, want unchanged %q", updated.LabelID, "label-123")
	}
}

func TestMemory_ListCarriesLabelID(t *testing.T) {
	r := NewMemoryRepository()
	n, err := r.Create(context.Background(), Note{Title: "Note"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	labelID := "label-123"
	if _, err := r.Update(context.Background(), n.ID, UpdateInput{LabelID: &labelID}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	list, err := r.List(context.Background(), ListFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || list[0].LabelID == nil || *list[0].LabelID != "label-123" {
		t.Fatalf("List = %+v, want one summary carrying label %q", list, "label-123")
	}
}
