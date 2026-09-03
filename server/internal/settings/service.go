package settings

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

// hhmm matches a 24-hour "HH:MM" wall-clock time.
var hhmm = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

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

	if input.Notifications != nil {
		if !hhmm.MatchString(input.Notifications.MorningTime) {
			return Settings{}, apperr.InvalidInput("morningTime must be in HH:MM (24-hour) format")
		}
		if input.Notifications.RecipientEmail != nil {
			if e := strings.TrimSpace(*input.Notifications.RecipientEmail); e != "" &&
				(!strings.Contains(e, "@") || strings.Contains(e, " ")) {
				return Settings{}, apperr.InvalidInput("recipientEmail must be a valid email address")
			}
		}
	}

	return s.repo.Update(ctx, input)
}
