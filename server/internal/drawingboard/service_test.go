package drawingboard

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

func ptr(s string) *string { return &s }

func assertNotFound(t *testing.T, err error) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != apperr.KindNotFound {
		t.Fatalf("err = %v, want apperr.NotFound", err)
	}
}

func assertInvalidInput(t *testing.T, err error) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != apperr.KindInvalidInput {
		t.Fatalf("err = %v, want apperr.InvalidInput", err)
	}
}

func TestService_CreateDefaultsNameToCreatedAt(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	b, err := svc.Create(context.Background(), CreateInput{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if b.Name != b.CreatedAt.Format(DefaultNameLayout) {
		t.Errorf("Name = %q, want formatted CreatedAt %q", b.Name, b.CreatedAt.Format(DefaultNameLayout))
	}
}

func TestService_CreateTrimsAndKeepsProvidedName(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	b, err := svc.Create(context.Background(), CreateInput{Name: ptr("  Wireframes  ")})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if b.Name != "Wireframes" {
		t.Errorf("Name = %q, want %q", b.Name, "Wireframes")
	}
}

func TestService_CreateSeedsEmptyScene(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	b, err := svc.Create(context.Background(), CreateInput{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	var scene map[string]any
	if err := json.Unmarshal(b.SceneData, &scene); err != nil {
		t.Fatalf("SceneData isn't valid JSON: %v", err)
	}
	if _, ok := scene["elements"]; !ok {
		t.Errorf("SceneData = %s, want an \"elements\" key", b.SceneData)
	}
}

func TestService_UpdateRedefaultsBlankNameToCreatedAt(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	b, err := svc.Create(context.Background(), CreateInput{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updated, err := svc.Update(context.Background(), b.ID, UpdateInput{Name: ptr("   ")})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != b.CreatedAt.Format(DefaultNameLayout) {
		t.Errorf("Name = %q, want formatted CreatedAt %q", updated.Name, b.CreatedAt.Format(DefaultNameLayout))
	}
}

func TestService_UpdateSceneDataOnlyLeavesNameUnchanged(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	b, err := svc.Create(context.Background(), CreateInput{Name: ptr("Sketch")})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	scene := json.RawMessage(`{"elements":[{"id":"a","type":"rectangle"}],"appState":{},"files":{}}`)
	updated, err := svc.Update(context.Background(), b.ID, UpdateInput{SceneData: scene})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "Sketch" {
		t.Errorf("Name = %q, want unchanged %q", updated.Name, "Sketch")
	}
	if string(updated.SceneData) != string(scene) {
		t.Errorf("SceneData = %s, want %s", updated.SceneData, scene)
	}
}

func TestService_UpdateRejectsMalformedSceneData(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	b, err := svc.Create(context.Background(), CreateInput{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = svc.Update(context.Background(), b.ID, UpdateInput{SceneData: json.RawMessage(`{not json`)})
	assertInvalidInput(t, err)
}

func TestService_UpdateRejectsNonObjectSceneData(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	b, err := svc.Create(context.Background(), CreateInput{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = svc.Update(context.Background(), b.ID, UpdateInput{SceneData: json.RawMessage(`"just a string"`)})
	assertInvalidInput(t, err)
}

func TestService_UpdateUnknownBoardReturnsNotFound(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	_, err := svc.Update(context.Background(), "missing", UpdateInput{Name: ptr("x")})
	assertNotFound(t, err)
}

func TestService_GetUnknownBoardReturnsNotFound(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	_, err := svc.Get(context.Background(), "missing")
	assertNotFound(t, err)
}

func TestService_DeleteThenGetReturnsNotFound(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	b, err := svc.Create(context.Background(), CreateInput{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := svc.Delete(context.Background(), b.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = svc.Get(context.Background(), b.ID)
	assertNotFound(t, err)
}

func TestService_ListOmitsDeletedAndSceneData(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	kept, err := svc.Create(context.Background(), CreateInput{Name: ptr("Keep")})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	gone, err := svc.Create(context.Background(), CreateInput{Name: ptr("Gone")})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := svc.Delete(context.Background(), gone.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	list, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || list[0].ID != kept.ID {
		t.Fatalf("List = %+v, want just %q", list, kept.ID)
	}
}
