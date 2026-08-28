package fitness

import "context"

// Repository is the persistence boundary for the whole fitness feature —
// cycles and the exercises / logs / foods that hang off them — following
// internal/upskill's "one interface per feature" shape. Update* methods
// take a full *UpdateInput (not a mutation closure) so the SQL layer can
// build a real SET clause. MemoryRepository is the in-memory stand-in used
// by tests; PostgresRepository is the real implementation.
type Repository interface {
	// CreateCycle archives the current active cycle (if any) and inserts
	// the new one as active, atomically.
	CreateCycle(ctx context.Context, cycle Cycle) (Cycle, error)
	ListCycles(ctx context.Context) ([]Cycle, error)
	GetCycle(ctx context.Context, id string) (Cycle, error)
	// GetActiveCycle returns the single active cycle; ok is false when
	// there is none.
	GetActiveCycle(ctx context.Context) (cycle Cycle, ok bool, err error)
	UpdateCycle(ctx context.Context, id string, input UpdateCycleInput) (Cycle, error)
	ArchiveCycle(ctx context.Context, id string) (Cycle, error)
	// ActivateCycle makes id the active cycle, archiving whichever cycle is
	// currently active (the "one active at a time" rule), atomically.
	ActivateCycle(ctx context.Context, id string) (Cycle, error)
	DeleteCycle(ctx context.Context, id string) error

	CreateExercise(ctx context.Context, exercise Exercise) (Exercise, error)
	// ListExercises returns every exercise for cycleID with TotalLogged
	// computed across its logs.
	ListExercises(ctx context.Context, cycleID string) ([]Exercise, error)
	GetExercise(ctx context.Context, id string) (Exercise, error)
	UpdateExercise(ctx context.Context, id string, input UpdateExerciseInput) (Exercise, error)
	DeleteExercise(ctx context.Context, id string) error

	UpsertExerciseLog(ctx context.Context, exerciseID, date string, quantity float64) (ExerciseLog, error)
	ListExerciseLogs(ctx context.Context, exerciseID string) ([]ExerciseLog, error)
	DeleteExerciseLog(ctx context.Context, id string) error

	UpsertWeightLog(ctx context.Context, cycleID, date string, weight float64) (WeightLog, error)
	ListWeightLogs(ctx context.Context, cycleID string) ([]WeightLog, error)
	DeleteWeightLog(ctx context.Context, id string) error

	// Foods are a single shared library, not cycle-scoped.
	CreateFood(ctx context.Context, food Food) (Food, error)
	ListFoods(ctx context.Context) ([]Food, error)
	GetFood(ctx context.Context, id string) (Food, error)
	UpdateFood(ctx context.Context, id string, input UpdateFoodInput) (Food, error)
	DeleteFood(ctx context.Context, id string) error

	CreateProteinLog(ctx context.Context, log ProteinLog) (ProteinLog, error)
	ListProteinLogs(ctx context.Context, cycleID string) ([]ProteinLog, error)
	DeleteProteinLog(ctx context.Context, id string) error
	ProteinDailyTotals(ctx context.Context, cycleID string) ([]ProteinDailyTotal, error)
}
