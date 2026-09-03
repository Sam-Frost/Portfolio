package scheduler

import (
	"testing"
	"time"

	"github.com/Sam-Frost/portfolio/internal/todo"
)

func TestDigestDue(t *testing.T) {
	loc := IST

	tests := []struct {
		name        string
		now         time.Time
		morningTime string
		wantDate    string
		wantDue     bool
	}{
		{
			name:        "before trigger time",
			now:         time.Date(2026, 9, 3, 6, 59, 0, 0, loc),
			morningTime: "07:00",
			wantDate:    "2026-09-03",
			wantDue:     false,
		},
		{
			name:        "exactly at trigger time",
			now:         time.Date(2026, 9, 3, 7, 0, 0, 0, loc),
			morningTime: "07:00",
			wantDate:    "2026-09-03",
			wantDue:     true,
		},
		{
			name:        "later in the day",
			now:         time.Date(2026, 9, 3, 23, 30, 0, 0, loc),
			morningTime: "07:00",
			wantDate:    "2026-09-03",
			wantDue:     true,
		},
		{
			name:        "UTC instant that is next day in IST",
			now:         time.Date(2026, 9, 3, 20, 0, 0, 0, time.UTC), // 01:30 IST on the 4th
			morningTime: "01:00",
			wantDate:    "2026-09-04",
			wantDue:     true,
		},
		{
			name:        "malformed morning time is never due",
			now:         time.Date(2026, 9, 3, 12, 0, 0, 0, loc),
			morningTime: "oops",
			wantDate:    "2026-09-03",
			wantDue:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, due := digestDue(tt.now, tt.morningTime, loc)
			if date != tt.wantDate || due != tt.wantDue {
				t.Fatalf("digestDue = (%q, %v), want (%q, %v)", date, due, tt.wantDate, tt.wantDue)
			}
		})
	}
}

func TestOverdueTodos(t *testing.T) {
	today := "2026-09-03"
	todos := []todo.Todo{
		{Name: "overdue", TargetDate: new("2026-09-01")},
		{Name: "due today", TargetDate: new("2026-09-03")},
		{Name: "future", TargetDate: new("2026-09-10")},
		{Name: "no date"},
		{Name: "done but overdue", TargetDate: new("2026-08-20"), Done: true},
	}

	got := overdueTodos(todos, today)
	if len(got) != 2 {
		t.Fatalf("got %d todos, want 2: %+v", len(got), got)
	}
	if got[0].Name != "overdue" || got[1].Name != "due today" {
		t.Fatalf("unexpected selection: %+v", got)
	}
}

func TestBuildDigestMessage(t *testing.T) {
	today := "2026-09-03"
	msg := buildDigestMessage([]todo.Todo{
		{Name: "file taxes", TargetDate: new("2026-09-01")},
		{Name: "call bank", TargetDate: new("2026-09-03")},
	}, today)

	if msg.Title != "2 todos due or overdue" {
		t.Fatalf("title = %q", msg.Title)
	}
	if msg.URL != "/todos" {
		t.Fatalf("url = %q", msg.URL)
	}
	wantBody := "• file taxes (due 2026-09-01)\n• call bank (due today)"
	if msg.Body != wantBody {
		t.Fatalf("body = %q, want %q", msg.Body, wantBody)
	}
}
