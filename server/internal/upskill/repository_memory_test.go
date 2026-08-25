package upskill

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

func mustCreateTopic(t *testing.T, repo *MemoryRepository, topic Topic) Topic {
	t.Helper()
	created, err := repo.CreateTopic(context.Background(), topic)
	if err != nil {
		t.Fatalf("CreateTopic(%+v): %v", topic, err)
	}
	return created
}

func mustCreateSubtopic(t *testing.T, repo *MemoryRepository, subtopic Subtopic) Subtopic {
	t.Helper()
	created, err := repo.CreateSubtopic(context.Background(), subtopic)
	if err != nil {
		t.Fatalf("CreateSubtopic(%+v): %v", subtopic, err)
	}
	return created
}

func TestMemoryRepository_ListTopicsComputesProgress(t *testing.T) {
	repo := NewMemoryRepository()
	topic := mustCreateTopic(t, repo, Topic{Name: "Go"})
	other := mustCreateTopic(t, repo, Topic{Name: "Rust"})

	s1 := mustCreateSubtopic(t, repo, Subtopic{TopicID: topic.ID, Name: "generics"})
	mustCreateSubtopic(t, repo, Subtopic{TopicID: topic.ID, Name: "channels"})

	done := true
	if _, err := repo.UpdateSubtopic(context.Background(), s1.ID, UpdateSubtopicInput{Done: &done}); err != nil {
		t.Fatalf("UpdateSubtopic: %v", err)
	}

	topics, err := repo.ListTopics(context.Background())
	if err != nil {
		t.Fatalf("ListTopics: %v", err)
	}

	var got, gotOther Topic
	for _, tpc := range topics {
		switch tpc.ID {
		case topic.ID:
			got = tpc
		case other.ID:
			gotOther = tpc
		}
	}
	if got.SubtopicCount != 2 {
		t.Errorf("SubtopicCount = %d, want 2", got.SubtopicCount)
	}
	if got.DoneCount != 1 {
		t.Errorf("DoneCount = %d, want 1", got.DoneCount)
	}
	if gotOther.SubtopicCount != 0 || gotOther.DoneCount != 0 {
		t.Errorf("other topic progress = %d/%d, want 0/0", gotOther.DoneCount, gotOther.SubtopicCount)
	}
}

func TestMemoryRepository_CreateSubtopicPersistsResources(t *testing.T) {
	repo := NewMemoryRepository()
	topic := mustCreateTopic(t, repo, Topic{Name: "Go"})

	label := "Docs"
	subtopic := mustCreateSubtopic(t, repo, Subtopic{
		TopicID: topic.ID,
		Name:    "generics",
		Resources: []Resource{
			{Label: &label, URL: "https://go.dev/doc"},
		},
	})

	if len(subtopic.Resources) != 1 {
		t.Fatalf("Resources = %+v, want 1 resource", subtopic.Resources)
	}
	if subtopic.Resources[0].SubtopicID != subtopic.ID {
		t.Errorf("Resources[0].SubtopicID = %q, want %q", subtopic.Resources[0].SubtopicID, subtopic.ID)
	}

	listed, err := repo.ListSubtopics(context.Background(), topic.ID)
	if err != nil {
		t.Fatalf("ListSubtopics: %v", err)
	}
	if len(listed) != 1 || len(listed[0].Resources) != 1 {
		t.Fatalf("ListSubtopics = %+v, want 1 subtopic with 1 resource", listed)
	}
}

func TestMemoryRepository_CreateSubtopicUnknownTopicReturnsInvalidInput(t *testing.T) {
	repo := NewMemoryRepository()

	_, err := repo.CreateSubtopic(context.Background(), Subtopic{TopicID: "missing", Name: "x"})
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != apperr.KindInvalidInput {
		t.Fatalf("err = %v, want apperr.InvalidInput", err)
	}
}

func TestMemoryRepository_DeleteTopicCascadesSubtopicsAndResources(t *testing.T) {
	repo := NewMemoryRepository()
	topic := mustCreateTopic(t, repo, Topic{Name: "Go"})
	label := "Docs"
	subtopic := mustCreateSubtopic(t, repo, Subtopic{
		TopicID:   topic.ID,
		Name:      "generics",
		Resources: []Resource{{Label: &label, URL: "https://go.dev/doc"}},
	})

	if err := repo.DeleteTopic(context.Background(), topic.ID); err != nil {
		t.Fatalf("DeleteTopic: %v", err)
	}

	if _, ok := repo.subtopics[subtopic.ID]; ok {
		t.Errorf("subtopic %q still present after topic delete", subtopic.ID)
	}
	if len(repo.resources) != 0 {
		t.Errorf("resources = %+v, want empty after topic delete", repo.resources)
	}
}

func TestMemoryRepository_UpdateSubtopicUnknownIDReturnsNotFound(t *testing.T) {
	repo := NewMemoryRepository()

	_, err := repo.UpdateSubtopic(context.Background(), "missing", UpdateSubtopicInput{})
	assertNotFound(t, err)
}

func TestMemoryRepository_DeleteResourceUnknownIDReturnsNotFound(t *testing.T) {
	repo := NewMemoryRepository()

	err := repo.DeleteResource(context.Background(), "missing")
	assertNotFound(t, err)
}
