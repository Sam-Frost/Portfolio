package worksession

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

func assertInvalidInput(t *testing.T, err error) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != apperr.KindInvalidInput {
		t.Fatalf("err = %v, want apperr.InvalidInput", err)
	}
}

func TestService_StartRejectsNonPositiveDuration(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	_, err := svc.Start(context.Background(), StartInput{PlannedMinutes: 0})
	assertInvalidInput(t, err)
}

func TestService_StartRejectsSecondRunningSession(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	ctx := context.Background()

	if _, err := svc.Start(ctx, StartInput{PlannedMinutes: 25}); err != nil {
		t.Fatalf("first Start: %v", err)
	}

	_, err := svc.Start(ctx, StartInput{PlannedMinutes: 10})
	assertInvalidInput(t, err)
}

func TestService_StartSucceedsAfterPriorSessionFinished(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	ctx := context.Background()

	first, err := svc.Start(ctx, StartInput{PlannedMinutes: 25})
	if err != nil {
		t.Fatalf("first Start: %v", err)
	}
	if _, err := svc.Cancel(ctx, first.ID, FinishBody{}); err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	if _, err := svc.Start(ctx, StartInput{PlannedMinutes: 10}); err != nil {
		t.Fatalf("second Start: %v", err)
	}
}

func TestService_CompleteRequiresGoalOrNote(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	ctx := context.Background()

	session, err := svc.Start(ctx, StartInput{PlannedMinutes: 25})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	_, err = svc.Complete(ctx, session.ID, FinishBody{Note: "   "})
	assertInvalidInput(t, err)
}

func TestService_CompleteRejectsUnknownOrNonRunningSession(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	_, err := svc.Complete(context.Background(), "does-not-exist", FinishBody{Note: "done"})
	assertInvalidInput(t, err)
}

func TestService_CompleteSetsStatusAndActualMinutes(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	session, err := svc.Start(ctx, StartInput{PlannedMinutes: 25})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	completed, err := svc.Complete(ctx, session.ID, FinishBody{Note: "wrote tests"})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if completed.Status != StatusCompleted {
		t.Errorf("Status = %v, want %v", completed.Status, StatusCompleted)
	}
	if completed.Note == nil || *completed.Note != "wrote tests" {
		t.Errorf("Note = %v, want %q", completed.Note, "wrote tests")
	}
	if completed.EndedAt == nil {
		t.Error("EndedAt = nil, want set")
	}
}

// A completed session ran its full planned duration by definition, so its
// end time is StartedAt + PlannedMinutes and its actual minutes equal the
// plan — regardless of how long after the timer elapsed the note is logged
// (the "time's up" dialog may sit open for hours). This is the bug the
// end-time-at-log-time behaviour had.
func TestService_CompleteEndsAtPlannedDurationNotLogTime(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	session, err := svc.Start(ctx, StartInput{PlannedMinutes: 25})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Simulate the timer having elapsed 90 minutes ago and the user only
	// now getting round to logging it.
	started := session.StartedAt.Add(-115 * time.Minute)
	repo.mu.Lock()
	s := repo.sessions[session.ID]
	s.StartedAt = started
	repo.sessions[session.ID] = s
	repo.mu.Unlock()

	completed, err := svc.Complete(ctx, session.ID, FinishBody{Note: "shipped it"})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if completed.ActualMinutes == nil || *completed.ActualMinutes != 25 {
		t.Errorf("ActualMinutes = %v, want 25 (the planned duration)", completed.ActualMinutes)
	}
	wantEnd := started.Add(25 * time.Minute)
	if completed.EndedAt == nil || !completed.EndedAt.Equal(wantEnd) {
		t.Errorf("EndedAt = %v, want %v (StartedAt + planned)", completed.EndedAt, wantEnd)
	}
}

func TestService_CompleteRecordsGoalsWithDoneFlags(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	ctx := context.Background()

	session, err := svc.Start(ctx, StartInput{
		PlannedMinutes: 25,
		Goals:          []string{"write handler", "write tests", "  "},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if len(session.Goals) != 2 {
		t.Fatalf("Start Goals = %v, want 2 (blank dropped)", session.Goals)
	}

	completed, err := svc.Complete(ctx, session.ID, FinishBody{
		Goals: []Goal{{Text: "write handler", Done: true}, {Text: "write tests", Done: false}},
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if len(completed.Goals) != 2 || !completed.Goals[0].Done || completed.Goals[1].Done {
		t.Errorf("Goals = %v, want [handler done, tests not done]", completed.Goals)
	}
}

func TestService_StartRejectsUnknownCategory(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	_, err := svc.Start(context.Background(), StartInput{PlannedMinutes: 25, Category: "hobby"})
	assertInvalidInput(t, err)
}

func TestService_StartDefaultsCategoryToProfessional(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	session, err := svc.Start(context.Background(), StartInput{PlannedMinutes: 25})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if session.Category != CategoryProfessional {
		t.Errorf("Category = %q, want %q", session.Category, CategoryProfessional)
	}
}

func TestService_CancelAllowsEmptyNote(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	ctx := context.Background()

	session, err := svc.Start(ctx, StartInput{PlannedMinutes: 25})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	cancelled, err := svc.Cancel(ctx, session.ID, FinishBody{})
	if err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if cancelled.Status != StatusCancelled {
		t.Errorf("Status = %v, want %v", cancelled.Status, StatusCancelled)
	}
	if cancelled.Note != nil {
		t.Errorf("Note = %v, want nil", cancelled.Note)
	}
	if cancelled.ActualMinutes == nil {
		t.Error("ActualMinutes = nil, want set even for a cancelled session")
	}
}

func TestService_CurrentReflectsRunningSession(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	ctx := context.Background()

	if _, ok, err := svc.Current(ctx); err != nil || ok {
		t.Fatalf("Current before Start = (ok=%v, err=%v), want (false, nil)", ok, err)
	}

	started, err := svc.Start(ctx, StartInput{PlannedMinutes: 25})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	current, ok, err := svc.Current(ctx)
	if err != nil || !ok {
		t.Fatalf("Current after Start = (ok=%v, err=%v), want (true, nil)", ok, err)
	}
	if current.ID != started.ID {
		t.Errorf("Current.ID = %q, want %q", current.ID, started.ID)
	}
}

// istTime builds a UTC time.Time from IST wall-clock components, for
// readable table-driven test cases below.
func istTime(year int, month time.Month, day, hour, min int) time.Time {
	return time.Date(year, month, day, hour, min, 0, 0, istLocation).UTC()
}

func TestSplitByISTDay_SameDaySessionIsNotSplit(t *testing.T) {
	start := istTime(2026, 8, 20, 10, 0)
	end := istTime(2026, 8, 20, 10, 45)

	got := splitByISTDay(start, end, 45)
	want := map[string]int{"2026-08-20": 45}
	if len(got) != len(want) || got["2026-08-20"] != 45 {
		t.Errorf("splitByISTDay same-day = %v, want %v", got, want)
	}
}

func TestSplitByISTDay_CrossMidnightSplitsProportionally(t *testing.T) {
	// 23:30 IST -> 00:30 IST the next day: 30 min on each side of 90 total.
	start := istTime(2026, 8, 20, 23, 30)
	end := istTime(2026, 8, 21, 0, 30)

	got := splitByISTDay(start, end, 60)
	if got["2026-08-20"] != 30 || got["2026-08-21"] != 30 {
		t.Errorf("splitByISTDay cross-midnight = %v, want {2026-08-20:30, 2026-08-21:30}", got)
	}
}

func TestSplitByISTDay_SumsExactlyToActualMinutesAcrossMultipleDays(t *testing.T) {
	// Spans three IST calendar days; uneven split shouldn't lose or gain a
	// minute to rounding.
	start := istTime(2026, 8, 20, 23, 40) // 20 min left in day 1
	end := istTime(2026, 8, 22, 0, 20)    // 20 min into day 3
	total := 100

	got := splitByISTDay(start, end, total)
	sum := 0
	for _, m := range got {
		sum += m
	}
	if sum != total {
		t.Errorf("sum of split minutes = %d, want %d (got %v)", sum, total, got)
	}
	if len(got) != 3 {
		t.Errorf("touched %d days, want 3 (got %v)", len(got), got)
	}
}

func TestSplitByISTDay_ZeroOrNegativeDurationYieldsNothing(t *testing.T) {
	start := istTime(2026, 8, 20, 10, 0)
	if got := splitByISTDay(start, start, 30); len(got) != 0 {
		t.Errorf("zero-duration split = %v, want empty", got)
	}
}

func TestService_DailySummarySplitsCrossMidnightSessionAcrossBothDays(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	session, err := svc.Start(ctx, StartInput{PlannedMinutes: 60})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	start := istTime(2026, 8, 20, 23, 30)
	end := istTime(2026, 8, 21, 0, 30)
	repo.mu.Lock()
	s := repo.sessions[session.ID]
	s.StartedAt = start
	repo.sessions[session.ID] = s
	repo.mu.Unlock()

	if _, err := svc.repo.Finish(ctx, session.ID, FinishInput{
		Status: StatusCompleted, EndedAt: end, ActualMinutes: 60,
	}); err != nil {
		t.Fatalf("Finish: %v", err)
	}

	from, _ := ParseISTDayStart("2026-08-20")
	to, _ := ParseISTDayStart("2026-08-22")
	summary, err := svc.DailySummary(ctx, from, to)
	if err != nil {
		t.Fatalf("DailySummary: %v", err)
	}

	byDate := make(map[string]int)
	for _, d := range summary {
		byDate[d.Date] = d.WorkedMinutes
	}
	if byDate["2026-08-20"] != 30 || byDate["2026-08-21"] != 30 {
		t.Errorf("DailySummary = %v, want {2026-08-20:30, 2026-08-21:30}", byDate)
	}
}

func TestService_DailySummaryIncludesCancelledSessions(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	session, err := svc.Start(ctx, StartInput{PlannedMinutes: 60})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Backdate so Cancel computes a non-zero elapsed duration — otherwise
	// (cancelled within the same minute) there's nothing to attribute and
	// this test would be checking the empty case instead.
	repo.mu.Lock()
	s := repo.sessions[session.ID]
	s.StartedAt = s.StartedAt.Add(-5 * time.Minute)
	repo.sessions[session.ID] = s
	repo.mu.Unlock()

	if _, err := svc.Cancel(ctx, session.ID, FinishBody{}); err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	today := time.Now().In(istLocation)
	dayStr := today.Format(DateLayout)
	from, _ := ParseISTDayStart(dayStr)
	to := from.AddDate(0, 0, 1)

	summary, err := svc.DailySummary(ctx, from, to)
	if err != nil {
		t.Fatalf("DailySummary: %v", err)
	}
	if len(summary) != 1 || summary[0].Date != dayStr {
		t.Fatalf("DailySummary = %v, want one entry for %s", summary, dayStr)
	}
	if summary[0].WorkedMinutes < 4 || summary[0].WorkedMinutes > 5 {
		t.Errorf("WorkedMinutes = %d, want ~5", summary[0].WorkedMinutes)
	}
}

func TestService_DailySummarySplitsMinutesByCategory(t *testing.T) {
	repo := NewMemoryRepository()
	svc := NewService(repo)
	ctx := context.Background()

	// One professional and one personal session on the same IST day, each
	// backdated so it has real elapsed minutes to attribute.
	for _, tc := range []struct {
		category Category
		ago      time.Duration
	}{
		{CategoryProfessional, 40 * time.Minute},
		{CategoryPersonal, 20 * time.Minute},
	} {
		session, err := svc.Start(ctx, StartInput{PlannedMinutes: 5, Category: tc.category})
		if err != nil {
			t.Fatalf("Start: %v", err)
		}
		repo.mu.Lock()
		s := repo.sessions[session.ID]
		s.StartedAt = s.StartedAt.Add(-tc.ago)
		repo.sessions[session.ID] = s
		repo.mu.Unlock()

		if _, err := svc.Complete(ctx, session.ID, FinishBody{Note: "done"}); err != nil {
			t.Fatalf("Complete: %v", err)
		}
	}

	dayStr := time.Now().In(istLocation).Format(DateLayout)
	from, _ := ParseISTDayStart(dayStr)
	summary, err := svc.DailySummary(ctx, from, from.AddDate(0, 0, 1))
	if err != nil {
		t.Fatalf("DailySummary: %v", err)
	}
	if len(summary) != 1 {
		t.Fatalf("DailySummary = %v, want one day", summary)
	}
	d := summary[0]
	if d.ProfessionalMinutes != 5 || d.PersonalMinutes != 5 || d.WorkedMinutes != 10 {
		t.Errorf("summary = %+v, want professional=5 personal=5 worked=10", d)
	}
}
