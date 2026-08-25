package settings

import (
	"context"
	"time"

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

	if input.TimeLeftClock != nil {
		if input.TimeLeftClock.Format != FormatWeeksDaysTime && input.TimeLeftClock.Format != FormatDaysTime {
			return Settings{}, apperr.InvalidInput("format must be weeks_days_time or days_time")
		}
		if input.TimeLeftClock.GoalDate != nil {
			if _, err := time.Parse(time.RFC3339, *input.TimeLeftClock.GoalDate); err != nil {
				return Settings{}, apperr.InvalidInput("goalDate must be a valid RFC3339 timestamp")
			}
		}
	}

	return s.repo.Update(ctx, input)
}
