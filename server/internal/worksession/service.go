package worksession

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

// istLocation is loaded once at package init. Every timestamp is still
// stored/compared as a UTC instant (time.Time is timezone-agnostic
// internally) — this is only used to decide which IST *calendar day* an
// instant falls on, per the "bucket by IST, store as UTC" convention.
var istLocation = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		// Fixed +05:30 fallback if the runtime image lacks tzdata — better
		// than the whole feature failing to boot over a missing database.
		return time.FixedZone("IST", 5*60*60+30*60)
	}
	return loc
}()

// ParseISTDayStart parses a "YYYY-MM-DD" date and returns the UTC instant
// corresponding to 00:00:00 IST on that date — the day-bucketing boundary
// used throughout this package.
func ParseISTDayStart(s string) (time.Time, error) {
	t, err := time.ParseInLocation(DateLayout, s, istLocation)
	if err != nil {
		return time.Time{}, apperr.InvalidInput("date must be in YYYY-MM-DD format")
	}
	return t, nil
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Start(ctx context.Context, input StartInput) (WorkSession, error) {
	if input.PlannedMinutes <= 0 {
		return WorkSession{}, apperr.InvalidInput("plannedMinutes must be positive")
	}

	_, running, err := s.repo.GetRunning(ctx)
	if err != nil {
		return WorkSession{}, err
	}
	if running {
		return WorkSession{}, apperr.InvalidInput("a session is already running")
	}

	return s.repo.Create(ctx, WorkSession{
		PlannedMinutes: input.PlannedMinutes,
		StartedAt:      time.Now().UTC(),
		Status:         StatusRunning,
	})
}

// Current returns the single running session, if any, for the frontend's
// floating timer to rehydrate from on page load.
func (s *Service) Current(ctx context.Context) (WorkSession, bool, error) {
	return s.repo.GetRunning(ctx)
}

func (s *Service) Complete(ctx context.Context, id string, input CompleteInput) (WorkSession, error) {
	note := strings.TrimSpace(input.Note)
	if note == "" {
		return WorkSession{}, apperr.InvalidInput("note is required to complete a session")
	}

	running, err := s.mustBeRunning(ctx, id)
	if err != nil {
		return WorkSession{}, err
	}

	now := time.Now().UTC()
	return s.repo.Finish(ctx, id, FinishInput{
		Status:        StatusCompleted,
		Note:          &note,
		EndedAt:       now,
		ActualMinutes: elapsedMinutes(running.StartedAt, now),
	})
}

func (s *Service) Cancel(ctx context.Context, id string, input CancelInput) (WorkSession, error) {
	var note *string
	if input.Note != nil {
		if trimmed := strings.TrimSpace(*input.Note); trimmed != "" {
			note = &trimmed
		}
	}

	running, err := s.mustBeRunning(ctx, id)
	if err != nil {
		return WorkSession{}, err
	}

	now := time.Now().UTC()
	return s.repo.Finish(ctx, id, FinishInput{
		Status:        StatusCancelled,
		Note:          note,
		EndedAt:       now,
		ActualMinutes: elapsedMinutes(running.StartedAt, now),
	})
}

// mustBeRunning confirms id is the currently running session, so
// Complete/Cancel can compute elapsed time from its StartedAt before
// handing off to Repository.Finish (which independently re-checks the
// 'running' status at the SQL layer against a race).
func (s *Service) mustBeRunning(ctx context.Context, id string) (WorkSession, error) {
	running, ok, err := s.repo.GetRunning(ctx)
	if err != nil {
		return WorkSession{}, err
	}
	if !ok || running.ID != id {
		return WorkSession{}, apperr.InvalidInput("session is not running")
	}
	return running, nil
}

func elapsedMinutes(start, end time.Time) int {
	m := int(end.Sub(start).Minutes())
	if m < 0 {
		m = 0
	}
	return m
}

// ListRange returns every session overlapping [from, to) for the calendar
// day-detail view (a cross-midnight session is returned once; the frontend
// decides, per day, whether it overlaps that day).
func (s *Service) ListRange(ctx context.Context, from, to time.Time) ([]WorkSession, error) {
	return s.repo.ListRange(ctx, from, to)
}

// DailySummary buckets actual worked minutes — completed and cancelled
// sessions alike, per the spec: even a cancelled session's partial run
// counts toward the day(s) it happened on — into IST calendar days over
// [from, to). A session whose StartedAt and EndedAt fall on different IST
// calendar dates has its ActualMinutes split proportionally between the
// days it touched at the midnight-IST boundary, so each day is credited
// only for the portion of the session that actually happened on it.
func (s *Service) DailySummary(ctx context.Context, from, to time.Time) ([]DailySummary, error) {
	sessions, err := s.repo.ListRange(ctx, from, to)
	if err != nil {
		return nil, err
	}

	totals := make(map[string]int)
	for _, sess := range sessions {
		if sess.EndedAt == nil || sess.ActualMinutes == nil {
			continue // still running — nothing finished to attribute yet
		}
		for date, minutes := range splitByISTDay(sess.StartedAt, *sess.EndedAt, *sess.ActualMinutes) {
			totals[date] += minutes
		}
	}

	dates := make([]string, 0, len(totals))
	for d := range totals {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	summaries := make([]DailySummary, 0, len(dates))
	for _, d := range dates {
		summaries = append(summaries, DailySummary{Date: d, WorkedMinutes: totals[d]})
	}
	return summaries, nil
}

// splitByISTDay divides actualMinutes across the IST calendar day(s)
// [start, end) touches, proportionally to how much of the [start, end)
// wall-clock span fell on each day. The common case — a session that
// doesn't cross midnight IST — is a single segment that gets all of
// actualMinutes with no rounding involved. When it does cross one or more
// midnights, every segment but the last is rounded down and the last
// absorbs the remainder, so the split always sums to exactly
// actualMinutes.
func splitByISTDay(start, end time.Time, actualMinutes int) map[string]int {
	result := make(map[string]int)
	if !end.After(start) || actualMinutes <= 0 {
		return result
	}

	type segment struct {
		date     string
		duration time.Duration
	}
	var segments []segment

	cursor := start
	for cursor.Before(end) {
		istCursor := cursor.In(istLocation)
		dayStart := time.Date(istCursor.Year(), istCursor.Month(), istCursor.Day(), 0, 0, 0, 0, istLocation)
		nextDayStart := dayStart.AddDate(0, 0, 1)

		segEnd := end
		if nextDayStart.Before(segEnd) {
			segEnd = nextDayStart
		}

		segments = append(segments, segment{date: dayStart.Format(DateLayout), duration: segEnd.Sub(cursor)})
		cursor = segEnd
	}

	if len(segments) == 1 {
		result[segments[0].date] = actualMinutes
		return result
	}

	total := end.Sub(start).Seconds()
	assigned := 0
	for i, seg := range segments {
		if i == len(segments)-1 {
			result[seg.date] += actualMinutes - assigned
			continue
		}
		minutes := int(float64(actualMinutes) * seg.duration.Seconds() / total)
		result[seg.date] += minutes
		assigned += minutes
	}
	return result
}
