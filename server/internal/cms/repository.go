package cms

import "context"

// Repository is the persistence boundary for all CMS content. The interface
// is shaped for the Postgres implementation (repository_postgres.go);
// repository_memory.go is an in-memory stand-in used by the service tests.
//
// List* methods return items already ordered by their Order field (then a
// stable tiebreak), so the service never re-sorts. Update* take a full
// Update*Input rather than a mutation closure, matching internal/todo.
type Repository interface {
	ListProjects(ctx context.Context) ([]Project, error)
	CreateProject(ctx context.Context, p Project) (Project, error)
	UpdateProject(ctx context.Context, id string, in UpdateProjectInput) (Project, error)
	DeleteProject(ctx context.Context, id string) error

	ListExperiences(ctx context.Context) ([]Experience, error)
	CreateExperience(ctx context.Context, e Experience) (Experience, error)
	UpdateExperience(ctx context.Context, id string, in UpdateExperienceInput) (Experience, error)
	DeleteExperience(ctx context.Context, id string) error

	ListBlogs(ctx context.Context) ([]Blog, error)
	CreateBlog(ctx context.Context, b Blog) (Blog, error)
	UpdateBlog(ctx context.Context, id string, in UpdateBlogInput) (Blog, error)
	DeleteBlog(ctx context.Context, id string) error

	GetSummary(ctx context.Context) (Summary, error)
	UpdateSummary(ctx context.Context, in UpdateSummaryInput) (Summary, error)

	// LatestPublication returns the most recent Publication of any status
	// (for showing "last publish failed" in the CMS) with its Snapshot
	// populated. LatestSuccessfulPublication returns the most recent one
	// that actually shipped — the baseline Status diffs the draft against.
	// Both return apperr.NotFound when there's no matching row.
	LatestPublication(ctx context.Context) (Publication, error)
	LatestSuccessfulPublication(ctx context.Context) (Publication, error)
	RecordPublication(ctx context.Context, p Publication) (Publication, error)
	ListPublications(ctx context.Context, limit int) ([]Publication, error)

	// NextMaxOrder returns 1 + the current highest Order across items of the
	// given section ("projects"|"experiences"|"blogs"), or 0 when empty, so
	// a newly created item sorts last.
	NextMaxOrder(ctx context.Context, section string) (int, error)
}
