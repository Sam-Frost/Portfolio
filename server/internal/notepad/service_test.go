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

	summaries, err := svc.List(context.Background())
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

	summaries, err := svc.List(context.Background())
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
