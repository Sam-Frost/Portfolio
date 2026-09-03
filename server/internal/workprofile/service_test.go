package workprofile

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

func mustInvalid(t *testing.T, err error) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != apperr.KindInvalidInput {
		t.Fatalf("err = %v, want apperr.InvalidInput", err)
	}
}

func seedTask(t *testing.T, svc *Service) (Tab, Task) {
	t.Helper()
	ctx := context.Background()
	tab, err := svc.CreateTab(ctx, CreateTabInput{Name: "Backend"})
	if err != nil {
		t.Fatalf("CreateTab: %v", err)
	}
	task, err := svc.CreateTask(ctx, tab.ID, CreateTaskInput{Name: "Fix the thing"})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	return tab, task
}

func TestUpdateTask_DoneRequiresJiraAck(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	_, task := seedTask(t, svc)
	ctx := context.Background()

	done := true
	_, err := svc.UpdateTask(ctx, task.ID, UpdateTaskInput{Done: &done})
	mustInvalid(t, err)

	no := false
	_, err = svc.UpdateTask(ctx, task.ID, UpdateTaskInput{Done: &done, JiraAcknowledged: &no})
	mustInvalid(t, err)

	yes := true
	got, err := svc.UpdateTask(ctx, task.ID, UpdateTaskInput{Done: &done, JiraAcknowledged: &yes})
	if err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	if !got.Done || got.CompletedAt == nil || !got.JiraAcknowledged {
		t.Fatalf("completed task looks wrong: %+v", got)
	}
}

func TestUpdateTask_UndoClearsAck(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	_, task := seedTask(t, svc)
	ctx := context.Background()

	done, yes := true, true
	if _, err := svc.UpdateTask(ctx, task.ID, UpdateTaskInput{Done: &done, JiraAcknowledged: &yes}); err != nil {
		t.Fatal(err)
	}
	notDone := false
	got, err := svc.UpdateTask(ctx, task.ID, UpdateTaskInput{Done: &notDone})
	if err != nil {
		t.Fatalf("UpdateTask undo: %v", err)
	}
	if got.Done || got.CompletedAt != nil || got.JiraAcknowledged {
		t.Fatalf("undone task should be reset: %+v", got)
	}
}

func TestOverview_BucketsByISTDate(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	ctx := context.Background()
	tab, _ := seedTask(t, svc)

	today := time.Now().In(IST).Format(TargetDateLayout)
	yesterday := time.Now().In(IST).AddDate(0, 0, -1).Format(TargetDateLayout)
	tomorrow := time.Now().In(IST).AddDate(0, 0, 1).Format(TargetDateLayout)

	mk := func(name, date string) {
		if _, err := svc.CreateTask(ctx, tab.ID, CreateTaskInput{Name: name, TargetDate: &date}); err != nil {
			t.Fatal(err)
		}
	}
	mk("due today", today)
	mk("overdue", yesterday)
	mk("future", tomorrow)

	ov, err := svc.Overview(ctx)
	if err != nil {
		t.Fatalf("Overview: %v", err)
	}
	if len(ov.DueToday) != 1 || ov.DueToday[0].Name != "due today" {
		t.Fatalf("dueToday = %+v", ov.DueToday)
	}
	if len(ov.Overdue) != 1 || ov.Overdue[0].Name != "overdue" {
		t.Fatalf("overdue = %+v", ov.Overdue)
	}
	if ov.DueToday[0].TabName != "Backend" {
		t.Fatalf("expected tab name on overview row, got %q", ov.DueToday[0].TabName)
	}
}

func TestCreateTab_RejectsBlankName(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	_, err := svc.CreateTab(context.Background(), CreateTabInput{Name: "  "})
	mustInvalid(t, err)
}
