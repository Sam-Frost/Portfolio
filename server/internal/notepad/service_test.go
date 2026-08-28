package notepad

import (
	"context"
	"strings"
	"testing"
)

func TestService_CreateDefaultsTitleToCreatedAt(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	n, err := svc.Create(context.Background(), CreateInput{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if n.Title != n.CreatedAt.Format(DefaultTitleLayout) {
		t.Errorf("Title = %q, want formatted CreatedAt %q", n.Title, n.CreatedAt.Format(DefaultTitleLayout))
	}
}

func TestService_CreateTrimsAndKeepsProvidedTitle(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	title := "  Shopping list  "
	n, err := svc.Create(context.Background(), CreateInput{Title: &title})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if n.Title != "Shopping list" {
		t.Errorf("Title = %q, want %q", n.Title, "Shopping list")
	}
}

func TestService_UpdateRedefaultsBlankTitleToCreatedAt(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	n, err := svc.Create(context.Background(), CreateInput{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	blank := "   "
	updated, err := svc.Update(context.Background(), n.ID, UpdateInput{Title: &blank})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Title != n.CreatedAt.Format(DefaultTitleLayout) {
		t.Errorf("Title = %q, want formatted CreatedAt %q", updated.Title, n.CreatedAt.Format(DefaultTitleLayout))
	}
}

func TestService_UpdateContentOnlyLeavesTitleUnchanged(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	title := "Notes"
	n, err := svc.Create(context.Background(), CreateInput{Title: &title})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	content := "<p>hello</p>"
	updated, err := svc.Update(context.Background(), n.ID, UpdateInput{ContentHTML: &content})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Title != "Notes" {
		t.Errorf("Title = %q, want unchanged %q", updated.Title, "Notes")
	}
	if !strings.Contains(updated.ContentHTML, "hello") {
		t.Errorf("ContentHTML = %q, want to contain %q", updated.ContentHTML, "hello")
	}
}

func TestService_ListOmitsContent(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	title := "Notes"
	n, err := svc.Create(context.Background(), CreateInput{Title: &title})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	content := "<p>secret body</p>"
	if _, err := svc.Update(context.Background(), n.ID, UpdateInput{ContentHTML: &content}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	summaries, err := svc.List(context.Background(), ListFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(summaries) != 1 || summaries[0].Title != "Notes" {
		t.Fatalf("List = %+v, want one summary titled %q", summaries, "Notes")
	}
}

func TestService_DeleteIsSoft(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	n, err := svc.Create(context.Background(), CreateInput{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.Delete(context.Background(), n.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := svc.Get(context.Background(), n.ID); err == nil {
		t.Errorf("Get after Delete = nil error, want not-found")
	}

	summaries, err := svc.List(context.Background(), ListFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(summaries) != 0 {
		t.Errorf("List after Delete = %+v, want none", summaries)
	}

	if err := svc.Delete(context.Background(), n.ID); err == nil {
		t.Errorf("second Delete = nil error, want not-found")
	}
}

func TestService_ArchiveMovesNoteBetweenListsAndClearsPin(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	n, err := svc.Create(context.Background(), CreateInput{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	pinned := true
	if _, err := svc.Update(context.Background(), n.ID, UpdateInput{Pinned: &pinned}); err != nil {
		t.Fatalf("Update pin: %v", err)
	}

	archived := true
	updated, err := svc.Update(context.Background(), n.ID, UpdateInput{Archived: &archived})
	if err != nil {
		t.Fatalf("Update archive: %v", err)
	}
	if !updated.Archived || updated.Pinned {
		t.Errorf("after archive: Archived=%v Pinned=%v, want true/false", updated.Archived, updated.Pinned)
	}

	active, err := svc.List(context.Background(), ListFilter{})
	if err != nil {
		t.Fatalf("List active: %v", err)
	}
	if len(active) != 0 {
		t.Errorf("active list = %+v, want empty", active)
	}

	archivedList, err := svc.List(context.Background(), ListFilter{Archived: true})
	if err != nil {
		t.Fatalf("List archived: %v", err)
	}
	if len(archivedList) != 1 || archivedList[0].ID != n.ID {
		t.Errorf("archived list = %+v, want the one note", archivedList)
	}
}

func TestService_ListSortsPinnedFirst(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	older, _ := svc.Create(context.Background(), CreateInput{})
	newer, _ := svc.Create(context.Background(), CreateInput{})

	pinned := true
	if _, err := svc.Update(context.Background(), older.ID, UpdateInput{Pinned: &pinned}); err != nil {
		t.Fatalf("Update pin: %v", err)
	}

	list, err := svc.List(context.Background(), ListFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 || list[0].ID != older.ID || list[1].ID != newer.ID {
		t.Errorf("list order = %+v, want pinned note (%s) first", list, older.ID)
	}
}
