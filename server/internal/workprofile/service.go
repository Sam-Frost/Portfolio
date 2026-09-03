package workprofile

import (
	"context"
	"strings"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

// IST is the timezone "due today" / "overdue" are measured in for the
// overview, resolved once like internal/diary and internal/scheduler.
var IST = mustLoadIST()

func mustLoadIST() *time.Location {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		panic("workprofile: failed to load Asia/Kolkata timezone: " + err.Error())
	}
	return loc
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// --- tabs ---

func (s *Service) ListTabs(ctx context.Context) ([]Tab, error) {
	return s.repo.ListTabs(ctx)
}

func (s *Service) CreateTab(ctx context.Context, input CreateTabInput) (Tab, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Tab{}, apperr.InvalidInput("name is required")
	}
	return s.repo.CreateTab(ctx, name)
}

func (s *Service) UpdateTab(ctx context.Context, id string, input UpdateTabInput) (Tab, error) {
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return Tab{}, apperr.InvalidInput("name is required")
	}
	return s.repo.UpdateTab(ctx, id, input)
}

func (s *Service) DeleteTab(ctx context.Context, id string) error {
	return s.repo.DeleteTab(ctx, id)
}

// --- tasks ---

func (s *Service) ListTasks(ctx context.Context, tabID string) ([]Task, error) {
	return s.repo.ListTasksByTab(ctx, tabID)
}

func (s *Service) CreateTask(ctx context.Context, tabID string, input CreateTaskInput) (Task, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Task{}, apperr.InvalidInput("name is required")
	}
	if err := validateTargetDate(input.TargetDate); err != nil {
		return Task{}, err
	}
	return s.repo.CreateTask(ctx, tabID, Task{
		Name:        name,
		Description: input.Description,
		TargetDate:  input.TargetDate,
	})
}

func (s *Service) UpdateTask(ctx context.Context, id string, input UpdateTaskInput) (Task, error) {
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return Task{}, apperr.InvalidInput("name is required")
	}
	if err := validateTargetDate(input.TargetDate); err != nil {
		return Task{}, err
	}
	// The Jira gate: a task can't be completed without confirming it was
	// logged in Jira. The dialog in the UI drives this, but the rule lives
	// here so it holds regardless of client.
	if input.Done != nil && *input.Done {
		if input.JiraAcknowledged == nil || !*input.JiraAcknowledged {
			return Task{}, apperr.InvalidInput("confirm the task was added to Jira before marking it done")
		}
	}
	return s.repo.UpdateTask(ctx, id, input)
}

func (s *Service) DeleteTask(ctx context.Context, id string) error {
	return s.repo.DeleteTask(ctx, id)
}

// Overview buckets every open task across all tabs into due-today and
// overdue (by IST calendar date). Tasks with no target date are omitted —
// they're not "due" anything.
func (s *Service) Overview(ctx context.Context) (Overview, error) {
	tasks, err := s.repo.ListOpenTasksWithTab(ctx)
	if err != nil {
		return Overview{}, err
	}

	today := time.Now().In(IST).Format(TargetDateLayout)
	out := Overview{DueToday: []TaskWithTab{}, Overdue: []TaskWithTab{}}
	for _, t := range tasks {
		if t.TargetDate == nil {
			continue
		}
		switch {
		case *t.TargetDate == today:
			out.DueToday = append(out.DueToday, t)
		case *t.TargetDate < today: // ISO dates compare lexicographically
			out.Overdue = append(out.Overdue, t)
		}
	}
	return out, nil
}

func validateTargetDate(targetDate *string) error {
	if targetDate == nil || *targetDate == "" {
		return nil
	}
	if _, err := time.Parse(TargetDateLayout, *targetDate); err != nil {
		return apperr.InvalidInput("targetDate must be in YYYY-MM-DD format")
	}
	return nil
}
