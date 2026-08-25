package upskill

import "context"

// Repository is the persistence boundary for topics, their subtopics, and
// each subtopic's resources. MemoryRepository (repository_memory.go) is an
// in-memory stand-in for PostgresRepository (repository_postgres.go);
// service/handler only depend on this interface. Update* methods take a
// full *UpdateInput (not a mutation closure) so the SQL implementation can
// build a real "SET" clause, matching internal/todo's Repository shape.
type Repository interface {
	CreateTopic(ctx context.Context, topic Topic) (Topic, error)
	// ListTopics returns every topic with SubtopicCount/DoneCount computed
	// across its subtopics.
	ListTopics(ctx context.Context) ([]Topic, error)
	GetTopic(ctx context.Context, id string) (Topic, error)
	UpdateTopic(ctx context.Context, id string, input UpdateTopicInput) (Topic, error)
	DeleteTopic(ctx context.Context, id string) error

	// CreateSubtopic persists the subtopic and its Resources atomically.
	CreateSubtopic(ctx context.Context, subtopic Subtopic) (Subtopic, error)
	// ListSubtopics returns every subtopic for topicID, each with its
	// Resources populated.
	ListSubtopics(ctx context.Context, topicID string) ([]Subtopic, error)
	UpdateSubtopic(ctx context.Context, id string, input UpdateSubtopicInput) (Subtopic, error)
	DeleteSubtopic(ctx context.Context, id string) error

	AddResource(ctx context.Context, subtopicID string, resource Resource) (Resource, error)
	DeleteResource(ctx context.Context, id string) error
}
