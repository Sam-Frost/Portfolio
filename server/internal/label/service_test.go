package label

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

	_, err := svc.Create(context.Background(), CreateInput{Name: "Work", Color: "chartreuse"})
	assertInvalidInput(t, err)
}

func TestService_CreateTrimsName(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	l, err := svc.Create(context.Background(), CreateInput{Name: "  Work  ", Color: "blue"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if l.Name != "Work" {
		t.Errorf("Name = %q, want %q", l.Name, "Work")
	}
}

func TestService_CreateRejectsDuplicateNameCaseInsensitive(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	if _, err := svc.Create(context.Background(), CreateInput{Name: "Work", Color: "blue"}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err := svc.Create(context.Background(), CreateInput{Name: "work", Color: "red"})
	assertInvalidInput(t, err)
}

func TestService_UpdateRejectsUnknownColor(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	l, err := svc.Create(context.Background(), CreateInput{Name: "Work", Color: "blue"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	bad := "chartreuse"
	_, err = svc.Update(context.Background(), l.ID, UpdateInput{Color: &bad})
	assertInvalidInput(t, err)
}
