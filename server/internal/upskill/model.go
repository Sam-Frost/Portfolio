package upskill

import "time"

type Topic struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	TargetDate    *string   `json:"targetDate"`
	DateAdded     time.Time `json:"dateAdded"`
	SubtopicCount int       `json:"subtopicCount"`
	DoneCount     int       `json:"doneCount"`
}

type Subtopic struct {
	ID         string     `json:"id"`
	TopicID    string     `json:"topicId"`
	Name       string     `json:"name"`
	TargetDate *string    `json:"targetDate"`
	Done       bool       `json:"done"`
	DateAdded  time.Time  `json:"dateAdded"`
	Resources  []Resource `json:"resources"`
}

type Resource struct {
	ID         string  `json:"id"`
	SubtopicID string  `json:"subtopicId"`
	Label      *string `json:"label"`
	URL        string  `json:"url"`
}

// TargetDateLayout is the wire format for TargetDate fields ("YYYY-MM-DD"),
// matching internal/todo's convention.
const TargetDateLayout = "2006-01-02"

type CreateTopicInput struct {
	Name       string  `json:"name"`
	TargetDate *string `json:"targetDate"`
}

// UpdateTopicInput is a partial update: nil fields are left unchanged. An
// empty-string TargetDate means "clear it" rather than "leave unchanged",
// matching internal/todo's UpdateInput convention.
type UpdateTopicInput struct {
	Name       *string `json:"name"`
	TargetDate *string `json:"targetDate"`
}

type CreateResourceInput struct {
	Label *string `json:"label"`
	URL   string  `json:"url"`
}

type CreateSubtopicInput struct {
	Name       string                `json:"name"`
	TargetDate *string               `json:"targetDate"`
	Resources  []CreateResourceInput `json:"resources"`
}

// UpdateSubtopicInput is a partial update: nil fields are left unchanged.
type UpdateSubtopicInput struct {
	Name       *string `json:"name"`
	TargetDate *string `json:"targetDate"`
	Done       *bool   `json:"done"`
}
