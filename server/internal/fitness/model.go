package fitness

import "time"

// DateLayout is the wire and DB format for every calendar-day field in this
// package ("YYYY-MM-DD"), interpreted as an IST calendar date — matching
// upskill.TargetDateLayout / diary.EntryDateLayout.
const DateLayout = "2006-01-02"

// CycleStatus is "active" or "archived". Exactly one cycle is active at a
// time (a partial unique index enforces it); CreateCycle archives the
// current active one.
type CycleStatus string

const (
	StatusActive   CycleStatus = "active"
	StatusArchived CycleStatus = "archived"
)

// Cycle is a training block — the project-like container every exercise,
// weigh-in, food, and protein log belongs to. WeightStart/WeightTarget/
// ProteinTarget are optional goals set on the cycle itself.
type Cycle struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	StartDate     string      `json:"startDate"`
	WeightStart   *float64    `json:"weightStart"`
	WeightTarget  *float64    `json:"weightTarget"`
	ProteinTarget *float64    `json:"proteinTarget"`
	Status        CycleStatus `json:"status"`
	CreatedAt     time.Time   `json:"createdAt"`
	ArchivedAt    *time.Time  `json:"archivedAt"`

	// Computed for list/detail views, not stored: how many exercises the
	// cycle has and its most recent weigh-in (nil if none logged yet).
	ExerciseCount int      `json:"exerciseCount"`
	LatestWeight  *float64 `json:"latestWeight"`
}

// Exercise is one tracked movement within a cycle, with an optional goal
// (target date + quantity). TotalLogged is computed (sum of its logs'
// quantities), not stored — like upskill.Topic.DoneCount.
type Exercise struct {
	ID           string    `json:"id"`
	CycleID      string    `json:"cycleId"`
	Name         string    `json:"name"`
	GoalDate     *string   `json:"goalDate"`
	GoalQuantity *float64  `json:"goalQuantity"`
	Unit         *string   `json:"unit"`
	CreatedAt    time.Time `json:"createdAt"`
	TotalLogged  float64   `json:"totalLogged"`
}

// ExerciseLog is one day's recorded quantity for an exercise. There is at
// most one per (exercise, date) — writes upsert.
type ExerciseLog struct {
	ID         string  `json:"id"`
	ExerciseID string  `json:"exerciseId"`
	LogDate    string  `json:"logDate"`
	Quantity   float64 `json:"quantity"`
}

// WeightLog is one day's weigh-in for a cycle. At most one per (cycle,
// date) — writes upsert.
type WeightLog struct {
	ID      string  `json:"id"`
	CycleID string  `json:"cycleId"`
	LogDate string  `json:"logDate"`
	Weight  float64 `json:"weight"`
}

// Food is a reusable entry in the shared food library: a name, a unit
// ("glass", "piece", "ml", "cup", "gm", ...), and the protein content per
// one of that unit. The library is not cycle-scoped — it's one list,
// edited in Settings and reused across every cycle.
type Food struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Unit           string    `json:"unit"`
	ProteinPerUnit float64   `json:"proteinPerUnit"`
	CreatedAt      time.Time `json:"createdAt"`
}

// ProteinLog is one logged serving: a food, a quantity (in that food's
// unit), and the resulting protein — snapshotted at log time so later
// edits to the food don't rewrite history. Many per day.
type ProteinLog struct {
	ID        string    `json:"id"`
	CycleID   string    `json:"cycleId"`
	FoodID    string    `json:"foodId"`
	LogDate   string    `json:"logDate"`
	Quantity  float64   `json:"quantity"`
	Protein   float64   `json:"protein"`
	CreatedAt time.Time `json:"createdAt"`
}

// ProteinDailyTotal is one day's summed protein intake, for the
// intake-vs-target and intake-over-time charts.
type ProteinDailyTotal struct {
	Date    string  `json:"date"`
	Protein float64 `json:"protein"`
}

// --- inputs ---

type CreateCycleInput struct {
	Name          string   `json:"name"`
	StartDate     string   `json:"startDate"`
	WeightStart   *float64 `json:"weightStart"`
	WeightTarget  *float64 `json:"weightTarget"`
	ProteinTarget *float64 `json:"proteinTarget"`
}

// UpdateCycleInput is a partial update: a nil field is left unchanged. A
// pointer to a zero value (e.g. WeightStart -> 0) is rejected by the
// service like any other non-positive weight, so "clear this goal" isn't
// expressible — matching how the frontend edit dialog always submits the
// full current set.
type UpdateCycleInput struct {
	Name          *string  `json:"name"`
	StartDate     *string  `json:"startDate"`
	WeightStart   *float64 `json:"weightStart"`
	WeightTarget  *float64 `json:"weightTarget"`
	ProteinTarget *float64 `json:"proteinTarget"`
}

type CreateExerciseInput struct {
	Name         string   `json:"name"`
	GoalDate     *string  `json:"goalDate"`
	GoalQuantity *float64 `json:"goalQuantity"`
	Unit         *string  `json:"unit"`
}

// UpdateExerciseInput is a partial update: nil fields are left unchanged.
// An empty-string GoalDate/Unit means "clear it" (matching
// upskill.UpdateTopicInput's TargetDate convention).
type UpdateExerciseInput struct {
	Name         *string  `json:"name"`
	GoalDate     *string  `json:"goalDate"`
	GoalQuantity *float64 `json:"goalQuantity"`
	Unit         *string  `json:"unit"`
}

// UpsertLogInput is the body of PUT .../logs and PUT .../weight-logs: a
// date plus the value for that date, replacing any existing row.
type UpsertExerciseLogInput struct {
	Date     string  `json:"date"`
	Quantity float64 `json:"quantity"`
}

type UpsertWeightLogInput struct {
	Date   string  `json:"date"`
	Weight float64 `json:"weight"`
}

type CreateFoodInput struct {
	Name           string  `json:"name"`
	Unit           string  `json:"unit"`
	ProteinPerUnit float64 `json:"proteinPerUnit"`
}

type UpdateFoodInput struct {
	Name           *string  `json:"name"`
	Unit           *string  `json:"unit"`
	ProteinPerUnit *float64 `json:"proteinPerUnit"`
}

type CreateProteinLogInput struct {
	FoodID   string  `json:"foodId"`
	Date     string  `json:"date"`
	Quantity float64 `json:"quantity"`
}
