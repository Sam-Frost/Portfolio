// Package scheduler is the process's single background timer. Everything
// else in the backend is evaluated lazily on an HTTP request; notifications
// are the exception — the morning todo digest and per-todo reminders have
// to fire on a clock with nobody watching.
//
// One ticker, one goroutine, started from cmd/main.go with a context tied
// to the shutdown signal. Safe to run on multiple instances: the digest
// dedups through notification_log's UNIQUE(kind, ist_date).
package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/Sam-Frost/portfolio/internal/notification"
	"github.com/Sam-Frost/portfolio/internal/reminder"
	"github.com/Sam-Frost/portfolio/internal/settings"
	"github.com/Sam-Frost/portfolio/internal/todo"
)

// IST is the timezone the "morning" of the digest is measured in, resolved
// once like internal/diary. A digest configured for 07:00 goes out at 07:00
// Asia/Kolkata regardless of the server's own clock zone.
var IST = mustLoadIST()

func mustLoadIST() *time.Location {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		panic("scheduler: failed to load Asia/Kolkata timezone: " + err.Error())
	}
	return loc
}

type Notifier interface {
	Notify(ctx context.Context, m notification.Message) error
	AlreadySent(ctx context.Context, kind, istDate string) (bool, error)
	RecordSent(ctx context.Context, kind, istDate string) error
}

type TodoLister interface {
	List(ctx context.Context, sortField todo.SortField, order todo.SortOrder, labelID *string) ([]todo.Todo, error)
}

type SettingsReader interface {
	Get(ctx context.Context) (settings.Settings, error)
}

type ReminderFirer interface {
	DueBefore(ctx context.Context, t time.Time) ([]reminder.Due, error)
	Settle(ctx context.Context, d reminder.Due, now time.Time) error
}

type Deps struct {
	Notifications Notifier
	Todos         TodoLister
	Settings      SettingsReader
	Reminders     ReminderFirer
}

type Scheduler struct {
	deps Deps
}

func New(deps Deps) *Scheduler { return &Scheduler{deps: deps} }

// Run blocks until ctx is cancelled, ticking once a minute. A minute is
// plenty of resolution for a daily digest and for reminders (whose UI
// granularity is minutes).
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	slog.Info("scheduler started")
	// Run one pass immediately so a deploy that happens to land right after
	// the digest time still sends today's digest.
	s.tick(ctx, time.Now())

	for {
		select {
		case <-ctx.Done():
			slog.Info("scheduler stopped")
			return
		case now := <-ticker.C:
			s.tick(ctx, now)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context, now time.Time) {
	s.maybeSendMorningDigest(ctx, now)
	s.fireDueReminders(ctx, now)
}

func (s *Scheduler) fireDueReminders(ctx context.Context, now time.Time) {
	if s.deps.Reminders == nil {
		return
	}

	due, err := s.deps.Reminders.DueBefore(ctx, now)
	if err != nil {
		slog.Error("scheduler: query due reminders failed", "err", err)
		return
	}

	for _, d := range due {
		msg := notification.Message{
			Title: "Reminder: " + d.TodoName,
			URL:   "/todos",
			Tag:   "reminder-" + d.TodoID,
		}
		if err := s.deps.Notifications.Notify(ctx, msg); err != nil {
			slog.Error("scheduler: send reminder failed", "err", err, "reminder", d.ID)
			// Still settle it — a delivery failure shouldn't wedge a
			// repeating reminder into firing every minute forever.
		}
		if err := s.deps.Reminders.Settle(ctx, d, now); err != nil {
			slog.Error("scheduler: settle reminder failed", "err", err, "reminder", d.ID)
		}
	}
}

func (s *Scheduler) maybeSendMorningDigest(ctx context.Context, now time.Time) {
	cfg, err := s.deps.Settings.Get(ctx)
	if err != nil {
		slog.Error("scheduler: get settings failed", "err", err)
		return
	}

	istDate, due := digestDue(now, cfg.Notifications.MorningTime, IST)
	if !due {
		return
	}

	sent, err := s.deps.Notifications.AlreadySent(ctx, notification.KindMorningDigest, istDate)
	if err != nil {
		slog.Error("scheduler: digest dedup check failed", "err", err)
		return
	}
	if sent {
		return
	}

	todos, err := s.deps.Todos.List(ctx, todo.SortByTargetDate, todo.SortAsc, nil)
	if err != nil {
		slog.Error("scheduler: list todos failed", "err", err)
		return
	}

	overdue := overdueTodos(todos, istDate)
	// Record the send even when there's nothing overdue, so the empty check
	// doesn't re-run every minute for the rest of the day.
	if err := s.deps.Notifications.RecordSent(ctx, notification.KindMorningDigest, istDate); err != nil {
		slog.Error("scheduler: record digest failed", "err", err)
		return
	}
	if len(overdue) == 0 {
		return
	}

	if err := s.deps.Notifications.Notify(ctx, buildDigestMessage(overdue, istDate)); err != nil {
		slog.Error("scheduler: send digest failed", "err", err)
	}
}
