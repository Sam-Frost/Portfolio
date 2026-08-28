package documentlabel

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

	_, err := svc.Create(context.Background(), CreateInput{Name: "Taxes", Color: "chartreuse"})
	assertInvalidInput(t, err)
}

func TestService_CreateTrimsName(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	l, err := svc.Create(context.Background(), CreateInput{Name: "  Taxes  ", Color: "blue"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if l.Name != "Taxes" {
		t.Errorf("Name = %q, want %q", l.Name, "Taxes")
	}
}

func TestService_CreateRejectsDuplicateNameCaseInsensitive(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	if _, err := svc.Create(context.Background(), CreateInput{Name: "Taxes", Color: "blue"}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err := svc.Create(context.Background(), CreateInput{Name: "taxes", Color: "red"})
	assertInvalidInput(t, err)
}
