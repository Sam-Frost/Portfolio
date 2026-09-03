package settings

import (
	"context"
	"errors"
	"testing"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

func TestService_UpdateRejectsNegativeHours(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	bad := -1.0
	_, err := svc.Update(context.Background(), UpdateInput{
		DailyWorkTracker: &DailyWorkTrackerInput{TotalWorkHoursRequired: &bad},
	})

	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != apperr.KindInvalidInput {
		t.Fatalf("err = %v, want apperr.InvalidInput", err)
	}
}

func TestService_UpdateSetsAndClearsHours(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	hours := 7.5
	s, err := svc.Update(context.Background(), UpdateInput{
		DailyWorkTracker: &DailyWorkTrackerInput{TotalWorkHoursRequired: &hours},
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if s.DailyWorkTracker.TotalWorkHoursRequired == nil || *s.DailyWorkTracker.TotalWorkHoursRequired != hours {
		t.Fatalf("TotalWorkHoursRequired = %v, want %v", s.DailyWorkTracker.TotalWorkHoursRequired, hours)
	}

	s, err = svc.Update(context.Background(), UpdateInput{
		DailyWorkTracker: &DailyWorkTrackerInput{TotalWorkHoursRequired: nil},
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if s.DailyWorkTracker.TotalWorkHoursRequired != nil {
		t.Fatalf("TotalWorkHoursRequired = %v, want nil", s.DailyWorkTracker.TotalWorkHoursRequired)
	}
}

func TestService_UpdateRejectsBadMorningTime(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	_, err := svc.Update(context.Background(), UpdateInput{
		Notifications: &NotificationsInput{MorningTime: "7am", EmailEnabled: true},
	})

	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != apperr.KindInvalidInput {
		t.Fatalf("err = %v, want apperr.InvalidInput", err)
	}
}

func TestService_UpdateRejectsBadRecipientEmail(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	bad := "not-an-email"
	_, err := svc.Update(context.Background(), UpdateInput{
		Notifications: &NotificationsInput{MorningTime: "07:00", RecipientEmail: &bad},
	})

	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != apperr.KindInvalidInput {
		t.Fatalf("err = %v, want apperr.InvalidInput", err)
	}
}

func TestService_UpdateSetsNotifications(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	email := "me@example.com"
	s, err := svc.Update(context.Background(), UpdateInput{
		Notifications: &NotificationsInput{
			RecipientEmail: &email, MorningTime: "08:30", EmailEnabled: true, PushEnabled: false,
		},
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if s.Notifications.RecipientEmail == nil || *s.Notifications.RecipientEmail != email {
		t.Fatalf("RecipientEmail = %v, want %q", s.Notifications.RecipientEmail, email)
	}
	if s.Notifications.MorningTime != "08:30" || !s.Notifications.EmailEnabled || s.Notifications.PushEnabled {
		t.Fatalf("unexpected notifications: %+v", s.Notifications)
	}
}

func TestService_UpdateRejectsInvalidTimeLeftFormat(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	_, err := svc.Update(context.Background(), UpdateInput{
		TimeLeftClock: &TimeLeftClockInput{Format: "not_a_format"},
	})

	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != apperr.KindInvalidInput {
		t.Fatalf("err = %v, want apperr.InvalidInput", err)
	}
}

func TestService_UpdateRejectsInvalidGoalDate(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	bad := "not-a-date"
	_, err := svc.Update(context.Background(), UpdateInput{
		TimeLeftClock: &TimeLeftClockInput{Format: FormatDaysTime, GoalDate: &bad},
	})

	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != apperr.KindInvalidInput {
		t.Fatalf("err = %v, want apperr.InvalidInput", err)
	}
}

func TestService_UpdateSetsAndClearsGoalDate(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	goalDate := "2026-12-31T00:00:00Z"
	s, err := svc.Update(context.Background(), UpdateInput{
		TimeLeftClock: &TimeLeftClockInput{Format: FormatWeeksDaysTime, GoalDate: &goalDate},
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if s.TimeLeftClock.GoalDate == nil || *s.TimeLeftClock.GoalDate != goalDate {
		t.Fatalf("GoalDate = %v, want %v", s.TimeLeftClock.GoalDate, goalDate)
	}
	if s.TimeLeftClock.Format != FormatWeeksDaysTime {
		t.Fatalf("Format = %v, want %v", s.TimeLeftClock.Format, FormatWeeksDaysTime)
	}

	s, err = svc.Update(context.Background(), UpdateInput{
		TimeLeftClock: &TimeLeftClockInput{Format: FormatDaysTime, GoalDate: nil},
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if s.TimeLeftClock.GoalDate != nil {
		t.Fatalf("GoalDate = %v, want nil", s.TimeLeftClock.GoalDate)
	}
	if s.TimeLeftClock.Format != FormatDaysTime {
		t.Fatalf("Format = %v, want %v", s.TimeLeftClock.Format, FormatDaysTime)
	}
}
