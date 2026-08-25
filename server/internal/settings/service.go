package settings

import (
	"context"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get(ctx context.Context) (Settings, error) {
	return s.repo.Get(ctx)
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (Settings, error) {
	if input.DailyWorkTracker != nil && input.DailyWorkTracker.TotalWorkHoursRequired != nil {
		if *input.DailyWorkTracker.TotalWorkHoursRequired < 0 {
			return Settings{}, apperr.InvalidInput("totalWorkHoursRequired must be >= 0")
		}
	}

	return s.repo.Update(ctx, input)
}
