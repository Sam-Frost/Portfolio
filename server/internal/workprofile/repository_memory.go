package workprofile

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

type MemoryRepository struct {
	mu    sync.Mutex
	tabs  map[string]Tab
	tasks map[string]Task
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{tabs: map[string]Tab{}, tasks: map[string]Task{}}
}

func (r *MemoryRepository) CreateTab(_ context.Context, name string) (Tab, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	pos := 0
	for _, t := range r.tabs {
		if t.Position >= pos {
			pos = t.Position + 1
		}
	}
	tab := Tab{ID: id.New(), Name: name, Position: pos, CreatedAt: time.Now().UTC()}
	r.tabs[tab.ID] = tab
	return tab, nil
}

func (r *MemoryRepository) ListTabs(_ context.Context) ([]Tab, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Tab, 0, len(r.tabs))
	for _, t := range r.tabs {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Position != out[j].Position {
			return out[i].Position < out[j].Position
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out, nil
}

func (r *MemoryRepository) UpdateTab(_ context.Context, tabID string, input UpdateTabInput) (Tab, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tabs[tabID]
	if !ok {
		return Tab{}, apperr.NotFound("tab not found")
	}
	if input.Name != nil {
		t.Name = *input.Name
	}
	if input.Position != nil {
		t.Position = *input.Position
	}
	r.tabs[tabID] = t
	return t, nil
}

func (r *MemoryRepository) DeleteTab(_ context.Context, tabID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tabs[tabID]; !ok {
		return apperr.NotFound("tab not found")
	}
	delete(r.tabs, tabID)
	for id, task := range r.tasks {
		if task.TabID == tabID {
			delete(r.tasks, id)
		}
	}
	return nil
}

func (r *MemoryRepository) CreateTask(_ context.Context, tabID string, task Task) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tabs[tabID]; !ok {
		return Task{}, apperr.NotFound("tab not found")
	}
	task.ID = id.New()
	task.TabID = tabID
	task.CreatedAt = time.Now().UTC()
	task.Done = false
	task.CompletedAt = nil
	task.JiraAcknowledged = false
	r.tasks[task.ID] = task
	return task, nil
}

func (r *MemoryRepository) ListTasksByTab(_ context.Context, tabID string) ([]Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Task, 0)
	for _, t := range r.tasks {
		if t.TabID == tabID {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (r *MemoryRepository) UpdateTask(_ context.Context, taskID string, input UpdateTaskInput) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tasks[taskID]
	if !ok {
		return Task{}, apperr.NotFound("task not found")
	}
	if input.Name != nil {
		t.Name = *input.Name
	}
	if input.Description != nil {
		t.Description = input.Description
	}
	if input.TargetDate != nil {
		if *input.TargetDate == "" {
			t.TargetDate = nil
		} else {
			td := *input.TargetDate
			t.TargetDate = &td
		}
	}
	if input.Done != nil {
		t.Done = *input.Done
		if *input.Done {
			now := time.Now().UTC()
			t.CompletedAt = &now
			t.JiraAcknowledged = true
		} else {
			t.CompletedAt = nil
			t.JiraAcknowledged = false
		}
	}
	r.tasks[taskID] = t
	return t, nil
}

func (r *MemoryRepository) DeleteTask(_ context.Context, taskID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tasks[taskID]; !ok {
		return apperr.NotFound("task not found")
	}
	delete(r.tasks, taskID)
	return nil
}

func (r *MemoryRepository) ListOpenTasksWithTab(_ context.Context) ([]TaskWithTab, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]TaskWithTab, 0)
	for _, t := range r.tasks {
		if t.Done {
			continue
		}
		out = append(out, TaskWithTab{Task: t, TabName: r.tabs[t.TabID].Name})
	}
	return out, nil
}
