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
