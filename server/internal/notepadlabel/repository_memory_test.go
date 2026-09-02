package notepadlabel

import (
	"context"
	"errors"
	"testing"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

func assertNotFound(t *testing.T, err error) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != apperr.KindNotFound {
		t.Fatalf("err = %v, want apperr.NotFound", err)
	}
}

func mustCreate(t *testing.T, r *MemoryRepository, name, color string) Label {
	t.Helper()
	l, err := r.Create(context.Background(), Label{Name: name, Color: color})
	if err != nil {
		t.Fatalf("Create(%q): %v", name, err)
	}
	return l
}

func TestMemory_CreateRejectsDuplicateNameCaseInsensitive(t *testing.T) {
	r := NewMemoryRepository()
	mustCreate(t, r, "Ideas", "blue")

	_, err := r.Create(context.Background(), Label{Name: "IDEAS", Color: "red"})
	assertInvalidInput(t, err)
}

func TestMemory_ListSortsByNameAscending(t *testing.T) {
	r := NewMemoryRepository()
	mustCreate(t, r, "Work", "blue")
	mustCreate(t, r, "Admin", "red")
	mustCreate(t, r, "Personal", "green")

	list, err := r.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	got := []string{list[0].Name, list[1].Name, list[2].Name}
	want := []string{"Admin", "Personal", "Work"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("List order = %v, want %v", got, want)
		}
	}
}

func TestMemory_UpdateChangesNameAndColor(t *testing.T) {
	r := NewMemoryRepository()
	l := mustCreate(t, r, "Ideas", "blue")

	name, color := "Thoughts", "purple"
	updated, err := r.Update(context.Background(), l.ID, UpdateInput{Name: &name, Color: &color})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "Thoughts" || updated.Color != "purple" {
		t.Errorf("updated = %+v, want name=Thoughts color=purple", updated)
	}
}

func TestMemory_UpdateTrimsName(t *testing.T) {
	r := NewMemoryRepository()
	l := mustCreate(t, r, "Ideas", "blue")

	name := "  Thoughts  "
	updated, err := r.Update(context.Background(), l.ID, UpdateInput{Name: &name})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "Thoughts" {
		t.Errorf("Name = %q, want trimmed %q", updated.Name, "Thoughts")
	}
}

// Renaming a label to its own current name (a no-op rename) must not trip
// the duplicate-name guard against itself.
func TestMemory_UpdateAllowsRenamingLabelToItsOwnName(t *testing.T) {
	r := NewMemoryRepository()
	l := mustCreate(t, r, "Ideas", "blue")

	name := "Ideas"
	if _, err := r.Update(context.Background(), l.ID, UpdateInput{Name: &name}); err != nil {
		t.Fatalf("Update to same name: %v", err)
	}
}

func TestMemory_UpdateRejectsNameTakenByAnotherLabel(t *testing.T) {
	r := NewMemoryRepository()
	mustCreate(t, r, "Ideas", "blue")
	other := mustCreate(t, r, "Work", "red")

	name := "ideas"
	_, err := r.Update(context.Background(), other.ID, UpdateInput{Name: &name})
	assertInvalidInput(t, err)
}

func TestMemory_UpdateUnknownLabelReturnsNotFound(t *testing.T) {
	r := NewMemoryRepository()
	name := "x"
	_, err := r.Update(context.Background(), "missing", UpdateInput{Name: &name})
	assertNotFound(t, err)
}

func TestMemory_DeleteRemovesLabel(t *testing.T) {
	r := NewMemoryRepository()
	l := mustCreate(t, r, "Ideas", "blue")

	if err := r.Delete(context.Background(), l.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	list, err := r.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("List after delete = %+v, want empty", list)
	}
}

func TestMemory_DeleteUnknownLabelReturnsNotFound(t *testing.T) {
	r := NewMemoryRepository()
	assertNotFound(t, r.Delete(context.Background(), "missing"))
}
