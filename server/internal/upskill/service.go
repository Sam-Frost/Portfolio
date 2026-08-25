package upskill

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// validateTargetDate treats an empty string as "clear the target date"
// (distinct from a nil pointer, which means "leave it unchanged" on
// update) rather than a malformed date, matching internal/todo's helper.
func validateTargetDate(targetDate *string) error {
	if targetDate == nil || *targetDate == "" {
		return nil
	}
	if _, err := time.Parse(TargetDateLayout, *targetDate); err != nil {
		return apperr.InvalidInput("targetDate must be in YYYY-MM-DD format")
	}
	return nil
}

func validateResourceURL(raw string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return apperr.InvalidInput("resource url is required")
	}
	parsed, err := url.ParseRequestURI(trimmed)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return apperr.InvalidInput("resource url must be an absolute http(s) URL")
	}
	return nil
}

func (s *Service) CreateTopic(ctx context.Context, input CreateTopicInput) (Topic, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Topic{}, apperr.InvalidInput("name is required")
	}
	if err := validateTargetDate(input.TargetDate); err != nil {
		return Topic{}, err
	}

	return s.repo.CreateTopic(ctx, Topic{Name: name, TargetDate: input.TargetDate})
}

func (s *Service) ListTopics(ctx context.Context) ([]Topic, error) {
	return s.repo.ListTopics(ctx)
}

func (s *Service) GetTopic(ctx context.Context, id string) (Topic, error) {
	return s.repo.GetTopic(ctx, id)
}

func (s *Service) UpdateTopic(ctx context.Context, id string, input UpdateTopicInput) (Topic, error) {
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return Topic{}, apperr.InvalidInput("name is required")
	}
	if err := validateTargetDate(input.TargetDate); err != nil {
		return Topic{}, err
	}

	return s.repo.UpdateTopic(ctx, id, input)
}

func (s *Service) DeleteTopic(ctx context.Context, id string) error {
	return s.repo.DeleteTopic(ctx, id)
}

func (s *Service) CreateSubtopic(ctx context.Context, topicID string, input CreateSubtopicInput) (Subtopic, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Subtopic{}, apperr.InvalidInput("name is required")
	}
	if err := validateTargetDate(input.TargetDate); err != nil {
		return Subtopic{}, err
	}

	resources := make([]Resource, 0, len(input.Resources))
	for _, res := range input.Resources {
		if err := validateResourceURL(res.URL); err != nil {
			return Subtopic{}, err
		}
		var label *string
		if trimmed := strings.TrimSpace(deref(res.Label)); trimmed != "" {
			label = &trimmed
		}
		resources = append(resources, Resource{Label: label, URL: strings.TrimSpace(res.URL)})
	}

	return s.repo.CreateSubtopic(ctx, Subtopic{
		TopicID:    topicID,
		Name:       name,
		TargetDate: input.TargetDate,
		Resources:  resources,
	})
}

func (s *Service) ListSubtopics(ctx context.Context, topicID string) ([]Subtopic, error) {
	return s.repo.ListSubtopics(ctx, topicID)
}

func (s *Service) UpdateSubtopic(ctx context.Context, id string, input UpdateSubtopicInput) (Subtopic, error) {
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return Subtopic{}, apperr.InvalidInput("name is required")
	}
	if err := validateTargetDate(input.TargetDate); err != nil {
		return Subtopic{}, err
	}

	return s.repo.UpdateSubtopic(ctx, id, input)
}

func (s *Service) DeleteSubtopic(ctx context.Context, id string) error {
	return s.repo.DeleteSubtopic(ctx, id)
}

func (s *Service) AddResource(ctx context.Context, subtopicID string, input CreateResourceInput) (Resource, error) {
	if err := validateResourceURL(input.URL); err != nil {
		return Resource{}, err
	}

	var label *string
	if trimmed := strings.TrimSpace(deref(input.Label)); trimmed != "" {
		label = &trimmed
	}

	return s.repo.AddResource(ctx, subtopicID, Resource{Label: label, URL: strings.TrimSpace(input.URL)})
}

func (s *Service) DeleteResource(ctx context.Context, id string) error {
	return s.repo.DeleteResource(ctx, id)
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
