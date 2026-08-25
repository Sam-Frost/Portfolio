package todo

import "time"

type Todo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	DateAdded   time.Time `json:"dateAdded"`
	TargetDate  *string   `json:"targetDate"`
	Done        bool      `json:"done"`
	LabelID     *string   `json:"labelId"`
}

// TargetDateLayout is the wire format for Todo.TargetDate ("YYYY-MM-DD").
const TargetDateLayout = "2006-01-02"

type CreateInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	TargetDate  *string `json:"targetDate"`
	LabelID     *string `json:"labelId"`
}

// UpdateInput is a partial update: nil fields are left unchanged. Built for
// a repository to translate directly into a SQL "SET" clause later — no
// mutation-closure indirection to unwind when that lands. As with
// TargetDate, an empty-string LabelID means "clear it" rather than "leave
// unchanged".
type UpdateInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	TargetDate  *string `json:"targetDate"`
	Done        *bool   `json:"done"`
	LabelID     *string `json:"labelId"`
}
