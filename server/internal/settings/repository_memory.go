package settings

import (
	"context"
	"sync"
)

type MemoryRepository struct {
	mu       sync.Mutex
	settings Settings
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{settings: Settings{
		TimeLeftClock: TimeLeftClock{Format: FormatWeeksDaysTime},
		Notifications: Notifications{MorningTime: "07:00", EmailEnabled: true, PushEnabled: true},
	}}
}

func (r *MemoryRepository) Get(_ context.Context) (Settings, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.settings, nil
}

func (r *MemoryRepository) Update(_ context.Context, input UpdateInput) (Settings, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if input.DailyWorkTracker != nil {
		r.settings.DailyWorkTracker.TotalWorkHoursRequired = input.DailyWorkTracker.TotalWorkHoursRequired
	}

	if input.TimeLeftClock != nil {
		r.settings.TimeLeftClock.GoalDate = input.TimeLeftClock.GoalDate
		r.settings.TimeLeftClock.Format = input.TimeLeftClock.Format
	}

	if input.Notifications != nil {
		r.settings.Notifications = Notifications{
			RecipientEmail: input.Notifications.RecipientEmail,
			MorningTime:    input.Notifications.MorningTime,
			EmailEnabled:   input.Notifications.EmailEnabled,
			PushEnabled:    input.Notifications.PushEnabled,
		}
	}

	return r.settings, nil
}
