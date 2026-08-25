package todo

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

	_, err := svc.Create(context.Background(), CreateInput{Name: "   "})
	assertInvalidInput(t, err)
}

func TestService_CreateTrimsName(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	todo, err := svc.Create(context.Background(), CreateInput{Name: "  Buy milk  "})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if todo.Name != "Buy milk" {
		t.Errorf("Name = %q, want %q", todo.Name, "Buy milk")
	}
}

func TestService_CreateRejectsMalformedTargetDate(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	bad := "25-08-2026"

	_, err := svc.Create(context.Background(), CreateInput{Name: "Buy milk", TargetDate: &bad})
	assertInvalidInput(t, err)
}

func TestService_UpdateRejectsBlankName(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	todo, err := svc.Create(context.Background(), CreateInput{Name: "Buy milk"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	blank := "   "
	_, err = svc.Update(context.Background(), todo.ID, UpdateInput{Name: &blank})
	assertInvalidInput(t, err)
}

func TestService_UpdateRejectsMalformedTargetDate(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	todo, err := svc.Create(context.Background(), CreateInput{Name: "Buy milk"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	bad := "25-08-2026"
	_, err = svc.Update(context.Background(), todo.ID, UpdateInput{TargetDate: &bad})
	assertInvalidInput(t, err)
}

func TestService_UpdateEmptyTargetDateClearsIt(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	due := "2026-09-01"
	todo, err := svc.Create(context.Background(), CreateInput{Name: "Buy milk", TargetDate: &due})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	empty := ""
	updated, err := svc.Update(context.Background(), todo.ID, UpdateInput{TargetDate: &empty})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.TargetDate != nil {
		t.Errorf("TargetDate = %v, want nil", updated.TargetDate)
	}
}
