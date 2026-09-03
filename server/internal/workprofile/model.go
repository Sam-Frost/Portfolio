// Package workprofile is the "Work Profile" domain area: user-created tabs
// (like Jira boards / workstreams), each holding todo-like tasks. It's a
// separate feature from internal/todo — tasks are grouped by tab and
// completing one is gated on a "logged in Jira?" acknowledgement — so it
// gets its own tables and package rather than overloading todos.
package workprofile

import "time"

// Tab groups a set of tasks. Position orders the tab bar (ascending).
type Tab struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"createdAt"`
}

// TargetDateLayout is the wire + DB format for Task.TargetDate, matching
// todo.TargetDateLayout.
const TargetDateLayout = "2006-01-02"

// Task mirrors todo.Todo, plus JiraAcknowledged: a task can only be marked
// done once the user has confirmed it was logged in Jira (enforced in the
// service, not just the UI). CompletedAt tracks done the same way todos do.
type Task struct {
	ID               string     `json:"id"`
	TabID            string     `json:"tabId"`
	Name             string     `json:"name"`
	Description      *string    `json:"description"`
	TargetDate       *string    `json:"targetDate"`
	Done             bool       `json:"done"`
	CompletedAt      *time.Time `json:"completedAt"`
	JiraAcknowledged bool       `json:"jiraAcknowledged"`
	CreatedAt        time.Time  `json:"createdAt"`
}

// TaskWithTab is a task carrying its tab's name, for the cross-tab overview.
type TaskWithTab struct {
	Task
	TabName string `json:"tabName"`
}

// Overview is the aggregate view across every tab.
type Overview struct {
	DueToday []TaskWithTab `json:"dueToday"`
	Overdue  []TaskWithTab `json:"overdue"`
}

type CreateTabInput struct {
	Name string `json:"name"`
}

// UpdateTabInput is a partial update: nil fields are left unchanged.
type UpdateTabInput struct {
	Name     *string `json:"name"`
	Position *int    `json:"position"`
}

type CreateTaskInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	TargetDate  *string `json:"targetDate"`
}

// UpdateTaskInput is a partial update. Setting Done to true requires
// JiraAcknowledged to also be true in the same request.
type UpdateTaskInput struct {
	Name             *string `json:"name"`
	Description      *string `json:"description"`
	TargetDate       *string `json:"targetDate"`
	Done             *bool   `json:"done"`
	JiraAcknowledged *bool   `json:"jiraAcknowledged"`
}
