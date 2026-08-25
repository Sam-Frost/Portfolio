package label

import (
	"context"
	"strings"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func validateColor(color string) error {
	if !AllowedColors[color] {
		return apperr.InvalidInput("color must be one of the preset label colors")
	}
	return nil
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Label, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Label{}, apperr.InvalidInput("name is required")
	}
	if err := validateColor(input.Color); err != nil {
		return Label{}, err
	}

	return s.repo.Create(ctx, Label{Name: name, Color: input.Color})
}

func (s *Service) List(ctx context.Context) ([]Label, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, id string, input UpdateInput) (Label, error) {
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return Label{}, apperr.InvalidInput("name is required")
	}
	if input.Color != nil {
		if err := validateColor(*input.Color); err != nil {
			return Label{}, err
		}
	}

	return s.repo.Update(ctx, id, input)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
