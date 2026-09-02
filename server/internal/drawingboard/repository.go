package drawingboard

import "context"

// Repository is the persistence boundary for drawing boards, shaped for the
// eventual Postgres implementation (see repository_postgres.go);
// repository_memory.go is the in-memory stand-in used by tests. Get exists
// separately from List because List deliberately omits SceneData.
type Repository interface {
	Create(ctx context.Context, b Board) (Board, error)
	List(ctx context.Context) ([]BoardSummary, error)
	Get(ctx context.Context, id string) (Board, error)
	Update(ctx context.Context, id string, input UpdateInput) (Board, error)
	Delete(ctx context.Context, id string) error
}
