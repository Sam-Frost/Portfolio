package todo

import (
	"context"
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
// update) rather than a malformed date.
func validateTargetDate(targetDate *string) error {
	if targetDate == nil || *targetDate == "" {
		return nil
	}
	if _, err := time.Parse(TargetDateLayout, *targetDate); err != nil {
		return apperr.InvalidInput("targetDate must be in YYYY-MM-DD format")
	}
	return nil
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Todo, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Todo{}, apperr.InvalidInput("name is required")
	}
	if err := validateTargetDate(input.TargetDate); err != nil {
		return Todo{}, err
	}

	return s.repo.Create(ctx, Todo{
		Name:        name,
		Description: input.Description,
		TargetDate:  input.TargetDate,
		LabelID:     input.LabelID,
	})
}

func (s *Service) List(ctx context.Context, sortField SortField, order SortOrder, labelID *string) ([]Todo, error) {
	return s.repo.List(ctx, sortField, order, labelID)
}

func (s *Service) CountActive(ctx context.Context) (int, error) {
	return s.repo.CountActive(ctx)
}

func (s *Service) Update(ctx context.Context, id string, input UpdateInput) (Todo, error) {
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return Todo{}, apperr.InvalidInput("name is required")
	}
	if err := validateTargetDate(input.TargetDate); err != nil {
		return Todo{}, err
	}

	return s.repo.Update(ctx, id, input)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
