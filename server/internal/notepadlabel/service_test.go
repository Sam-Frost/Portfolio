package notepadlabel

import (
	"context"
	"errors"
	"testing"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

func assertInvalidInput(t *testing.T, err error) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != apperr.KindInvalidInput {
		t.Fatalf("err = %v, want apperr.InvalidInput", err)
	}
}

func TestService_CreateRejectsBlankName(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	_, err := svc.Create(context.Background(), CreateInput{Name: "   ", Color: "red"})
	assertInvalidInput(t, err)
}

func TestService_CreateRejectsUnknownColor(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	_, err := svc.Create(context.Background(), CreateInput{Name: "Ideas", Color: "chartreuse"})
	assertInvalidInput(t, err)
}

func TestService_CreateTrimsName(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	l, err := svc.Create(context.Background(), CreateInput{Name: "  Ideas  ", Color: "blue"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if l.Name != "Ideas" {
		t.Errorf("Name = %q, want %q", l.Name, "Ideas")
	}
}

func TestService_CreateRejectsDuplicateNameCaseInsensitive(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	if _, err := svc.Create(context.Background(), CreateInput{Name: "Ideas", Color: "blue"}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err := svc.Create(context.Background(), CreateInput{Name: "ideas", Color: "red"})
	assertInvalidInput(t, err)
}

func TestService_UpdateRejectsUnknownColor(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	l, err := svc.Create(context.Background(), CreateInput{Name: "Ideas", Color: "blue"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	bad := "chartreuse"
	_, err = svc.Update(context.Background(), l.ID, UpdateInput{Color: &bad})
	assertInvalidInput(t, err)
}

func TestService_UpdateRejectsBlankName(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	l, err := svc.Create(context.Background(), CreateInput{Name: "Ideas", Color: "blue"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	blank := "   "
	_, err = svc.Update(context.Background(), l.ID, UpdateInput{Name: &blank})
	assertInvalidInput(t, err)
}

func TestService_UpdateRecolorsLabel(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	l, err := svc.Create(context.Background(), CreateInput{Name: "Ideas", Color: "blue"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	color := "green"
	updated, err := svc.Update(context.Background(), l.ID, UpdateInput{Color: &color})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Color != "green" || updated.Name != "Ideas" {
		t.Errorf("updated = %+v, want color=green name=Ideas", updated)
	}
}

func TestService_DeleteRemovesLabelFromList(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	l, err := svc.Create(context.Background(), CreateInput{Name: "Ideas", Color: "blue"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.Delete(context.Background(), l.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	list, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("List after delete = %+v, want empty", list)
	}
}
