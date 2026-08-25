package upskill

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

func TestService_CreateTopicRejectsBlankName(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	_, err := svc.CreateTopic(context.Background(), CreateTopicInput{Name: "   "})
	assertInvalidInput(t, err)
}

func TestService_CreateTopicTrimsName(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	topic, err := svc.CreateTopic(context.Background(), CreateTopicInput{Name: "  Go  "})
	if err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}
	if topic.Name != "Go" {
		t.Errorf("Name = %q, want %q", topic.Name, "Go")
	}
}

func TestService_CreateTopicRejectsMalformedTargetDate(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	bad := "25-08-2026"

	_, err := svc.CreateTopic(context.Background(), CreateTopicInput{Name: "Go", TargetDate: &bad})
	assertInvalidInput(t, err)
}

func TestService_UpdateTopicEmptyTargetDateClearsIt(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	due := "2026-09-01"
	topic, err := svc.CreateTopic(context.Background(), CreateTopicInput{Name: "Go", TargetDate: &due})
	if err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}

	empty := ""
	updated, err := svc.UpdateTopic(context.Background(), topic.ID, UpdateTopicInput{TargetDate: &empty})
	if err != nil {
		t.Fatalf("UpdateTopic: %v", err)
	}
	if updated.TargetDate != nil {
		t.Errorf("TargetDate = %v, want nil", updated.TargetDate)
	}
}

func TestService_CreateSubtopicRejectsBlankName(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	topic, err := svc.CreateTopic(context.Background(), CreateTopicInput{Name: "Go"})
	if err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}

	_, err = svc.CreateSubtopic(context.Background(), topic.ID, CreateSubtopicInput{Name: "  "})
	assertInvalidInput(t, err)
}

func TestService_CreateSubtopicRejectsNonHTTPResourceURL(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	topic, err := svc.CreateTopic(context.Background(), CreateTopicInput{Name: "Go"})
	if err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}

	_, err = svc.CreateSubtopic(context.Background(), topic.ID, CreateSubtopicInput{
		Name:      "generics",
		Resources: []CreateResourceInput{{URL: "not-a-url"}},
	})
	assertInvalidInput(t, err)
}

func TestService_CreateSubtopicTrimsResourceLabelAndURL(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	topic, err := svc.CreateTopic(context.Background(), CreateTopicInput{Name: "Go"})
	if err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}

	label := "  Docs  "
	subtopic, err := svc.CreateSubtopic(context.Background(), topic.ID, CreateSubtopicInput{
		Name:      "generics",
		Resources: []CreateResourceInput{{Label: &label, URL: "  https://go.dev/doc  "}},
	})
	if err != nil {
		t.Fatalf("CreateSubtopic: %v", err)
	}
	if len(subtopic.Resources) != 1 {
		t.Fatalf("Resources = %+v, want 1", subtopic.Resources)
	}
	if got := subtopic.Resources[0]; *got.Label != "Docs" || got.URL != "https://go.dev/doc" {
		t.Errorf("Resource = %+v, want trimmed label/url", got)
	}
}

func TestService_AddResourceRejectsBlankURL(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	topic, err := svc.CreateTopic(context.Background(), CreateTopicInput{Name: "Go"})
	if err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}
	subtopic, err := svc.CreateSubtopic(context.Background(), topic.ID, CreateSubtopicInput{Name: "generics"})
	if err != nil {
		t.Fatalf("CreateSubtopic: %v", err)
	}

	_, err = svc.AddResource(context.Background(), subtopic.ID, CreateResourceInput{URL: "   "})
	assertInvalidInput(t, err)
}
