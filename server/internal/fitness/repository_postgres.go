package fitness

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// cycleSelect carries the stored columns plus two computed ones —
// exercise_count and latest_weight (most recent weigh-in) — for the list
// and detail views. Callers append their own WHERE / ORDER BY.
const cycleSelect = `
	SELECT c.id, c.name, c.start_date, c.weight_start, c.weight_target, c.protein_target,
	       c.status, c.created_at, c.archived_at,
	       COUNT(e.id) AS exercise_count,
	       (SELECT w.weight FROM fitness_weight_logs w
	          WHERE w.cycle_id = c.id ORDER BY w.log_date DESC LIMIT 1) AS latest_weight
	FROM fitness_cycles c
	LEFT JOIN fitness_exercises e ON e.cycle_id = c.id
`

func scanCycle(row interface{ Scan(...any) error }) (Cycle, error) {
	var c Cycle
	err := row.Scan(
		&c.ID, &c.Name, &c.StartDate, &c.WeightStart, &c.WeightTarget, &c.ProteinTarget,
		&c.Status, &c.CreatedAt, &c.ArchivedAt, &c.ExerciseCount, &c.LatestWeight,
	)
	return c, err
}

// --- cycles ---

func (r *PostgresRepository) CreateCycle(ctx context.Context, c Cycle) (Cycle, error) {
	c.ID = id.New()
	c.CreatedAt = time.Now().UTC()
	c.Status = StatusActive

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Cycle{}, apperr.Internal("failed to create cycle")
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`UPDATE fitness_cycles SET status = $1, archived_at = $2 WHERE status = $3`,
		StatusArchived, c.CreatedAt, StatusActive,
	); err != nil {
		return Cycle{}, apperr.Internal("failed to archive current cycle")
	}

	const q = `INSERT INTO fitness_cycles (id, name, start_date, weight_start, weight_target, protein_target, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	if _, err := tx.ExecContext(ctx, q,
		c.ID, c.Name, c.StartDate, c.WeightStart, c.WeightTarget, c.ProteinTarget, c.Status, c.CreatedAt,
	); err != nil {
		return Cycle{}, apperr.Internal("failed to create cycle")
	}

	if err := tx.Commit(); err != nil {
		return Cycle{}, apperr.Internal("failed to create cycle")
	}
	return c, nil
}

func (r *PostgresRepository) ListCycles(ctx context.Context) ([]Cycle, error) {
	// Active first, then most-recently-created archived.
	const q = cycleSelect + `
		GROUP BY c.id
		ORDER BY (c.status = 'active') DESC, c.created_at DESC`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, apperr.Internal("failed to list cycles")
	}
	defer rows.Close()

	cycles := make([]Cycle, 0)
	for rows.Next() {
		c, err := scanCycle(rows)
		if err != nil {
			return nil, apperr.Internal("failed to scan cycle")
		}
		cycles = append(cycles, c)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list cycles")
	}
	return cycles, nil
}

func (r *PostgresRepository) GetCycle(ctx context.Context, cycleID string) (Cycle, error) {
	c, err := scanCycle(r.db.QueryRowContext(ctx, cycleSelect+` WHERE c.id = $1 GROUP BY c.id`, cycleID))
	if errors.Is(err, sql.ErrNoRows) {
		return Cycle{}, apperr.NotFound("cycle not found")
	}
	if err != nil {
		return Cycle{}, apperr.Internal("failed to get cycle")
	}
	return c, nil
}

func (r *PostgresRepository) GetActiveCycle(ctx context.Context) (Cycle, bool, error) {
	c, err := scanCycle(r.db.QueryRowContext(ctx, cycleSelect+` WHERE c.status = 'active' GROUP BY c.id`))
	if errors.Is(err, sql.ErrNoRows) {
		return Cycle{}, false, nil
	}
	if err != nil {
		return Cycle{}, false, apperr.Internal("failed to get active cycle")
	}
	return c, true, nil
}

func (r *PostgresRepository) UpdateCycle(ctx context.Context, cycleID string, input UpdateCycleInput) (Cycle, error) {
	sets := make([]string, 0, 5)
	args := make([]any, 0, 6)
	argN := 1
	add := func(col string, val any) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, argN))
		args = append(args, val)
		argN++
	}

	if input.Name != nil {
		add("name", strings.TrimSpace(*input.Name))
	}
	if input.StartDate != nil {
		add("start_date", *input.StartDate)
	}
	if input.WeightStart != nil {
		add("weight_start", *input.WeightStart)
	}
	if input.WeightTarget != nil {
		add("weight_target", *input.WeightTarget)
	}
	if input.ProteinTarget != nil {
		add("protein_target", *input.ProteinTarget)
	}

	if len(sets) > 0 {
		args = append(args, cycleID)
		q := fmt.Sprintf("UPDATE fitness_cycles SET %s WHERE id = $%d", strings.Join(sets, ", "), argN)
		res, err := r.db.ExecContext(ctx, q, args...)
		if err != nil {
			return Cycle{}, apperr.Internal("failed to update cycle")
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return Cycle{}, apperr.NotFound("cycle not found")
		}
	}
	return r.GetCycle(ctx, cycleID)
}

func (r *PostgresRepository) ArchiveCycle(ctx context.Context, cycleID string) (Cycle, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE fitness_cycles SET status = $1, archived_at = $2 WHERE id = $3 AND status = $4`,
		StatusArchived, time.Now().UTC(), cycleID, StatusActive,
	)
	if err != nil {
		return Cycle{}, apperr.Internal("failed to archive cycle")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		// Either the cycle doesn't exist or it's already archived — the
		// GetCycle below turns the first case into a NotFound and the
		// second into a harmless no-op returning the current row.
		return r.GetCycle(ctx, cycleID)
	}
	return r.GetCycle(ctx, cycleID)
}

func (r *PostgresRepository) ActivateCycle(ctx context.Context, cycleID string) (Cycle, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Cycle{}, apperr.Internal("failed to activate cycle")
	}
	defer tx.Rollback()

	// Archive whichever cycle is currently active, except this one.
	if _, err := tx.ExecContext(ctx,
		`UPDATE fitness_cycles SET status = $1, archived_at = $2 WHERE status = $3 AND id <> $4`,
		StatusArchived, time.Now().UTC(), StatusActive, cycleID,
	); err != nil {
		return Cycle{}, apperr.Internal("failed to archive current cycle")
	}

	res, err := tx.ExecContext(ctx,
		`UPDATE fitness_cycles SET status = $1, archived_at = NULL WHERE id = $2`,
		StatusActive, cycleID,
	)
	if err != nil {
		return Cycle{}, apperr.Internal("failed to activate cycle")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return Cycle{}, apperr.NotFound("cycle not found")
	}

	if err := tx.Commit(); err != nil {
		return Cycle{}, apperr.Internal("failed to activate cycle")
	}
	return r.GetCycle(ctx, cycleID)
}

func (r *PostgresRepository) DeleteCycle(ctx context.Context, cycleID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM fitness_cycles WHERE id = $1`, cycleID)
	if err != nil {
		return apperr.Internal("failed to delete cycle")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return apperr.NotFound("cycle not found")
	}
	return nil
}

// --- exercises ---

func (r *PostgresRepository) CreateExercise(ctx context.Context, e Exercise) (Exercise, error) {
	e.ID = id.New()
	e.CreatedAt = time.Now().UTC()

	const q = `INSERT INTO fitness_exercises (id, cycle_id, name, goal_date, goal_quantity, unit, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	if _, err := r.db.ExecContext(ctx, q, e.ID, e.CycleID, e.Name, e.GoalDate, e.GoalQuantity, e.Unit, e.CreatedAt); err != nil {
		if isForeignKeyViolation(err) {
			return Exercise{}, apperr.InvalidInput("cycle not found")
		}
		return Exercise{}, apperr.Internal("failed to create exercise")
	}
	return e, nil
}

const exerciseSelect = `
	SELECT e.id, e.cycle_id, e.name, e.goal_date, e.goal_quantity, e.unit, e.created_at,
	       COALESCE(SUM(l.quantity), 0) AS total_logged
	FROM fitness_exercises e
	LEFT JOIN fitness_exercise_logs l ON l.exercise_id = e.id
`

func scanExercise(row interface{ Scan(...any) error }) (Exercise, error) {
	var e Exercise
	err := row.Scan(&e.ID, &e.CycleID, &e.Name, &e.GoalDate, &e.GoalQuantity, &e.Unit, &e.CreatedAt, &e.TotalLogged)
	return e, err
}

func (r *PostgresRepository) ListExercises(ctx context.Context, cycleID string) ([]Exercise, error) {
	q := exerciseSelect + ` WHERE e.cycle_id = $1 GROUP BY e.id ORDER BY e.created_at DESC`
	rows, err := r.db.QueryContext(ctx, q, cycleID)
	if err != nil {
		return nil, apperr.Internal("failed to list exercises")
	}
	defer rows.Close()

	exercises := make([]Exercise, 0)
	for rows.Next() {
		e, err := scanExercise(rows)
		if err != nil {
			return nil, apperr.Internal("failed to scan exercise")
		}
		exercises = append(exercises, e)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list exercises")
	}
	return exercises, nil
}

func (r *PostgresRepository) GetExercise(ctx context.Context, exerciseID string) (Exercise, error) {
	q := exerciseSelect + ` WHERE e.id = $1 GROUP BY e.id`
	e, err := scanExercise(r.db.QueryRowContext(ctx, q, exerciseID))
	if errors.Is(err, sql.ErrNoRows) {
		return Exercise{}, apperr.NotFound("exercise not found")
	}
	if err != nil {
		return Exercise{}, apperr.Internal("failed to get exercise")
	}
	return e, nil
}

func (r *PostgresRepository) UpdateExercise(ctx context.Context, exerciseID string, input UpdateExerciseInput) (Exercise, error) {
	sets := make([]string, 0, 4)
	args := make([]any, 0, 5)
	argN := 1
	add := func(col string, val any) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, argN))
		args = append(args, val)
		argN++
	}

	if input.Name != nil {
		add("name", strings.TrimSpace(*input.Name))
	}
	if input.GoalDate != nil {
		if *input.GoalDate == "" {
			add("goal_date", nil)
		} else {
			add("goal_date", *input.GoalDate)
		}
	}
	if input.GoalQuantity != nil {
		add("goal_quantity", *input.GoalQuantity)
	}
	if input.Unit != nil {
		if strings.TrimSpace(*input.Unit) == "" {
			add("unit", nil)
		} else {
			add("unit", strings.TrimSpace(*input.Unit))
		}
	}

	if len(sets) > 0 {
		args = append(args, exerciseID)
		q := fmt.Sprintf("UPDATE fitness_exercises SET %s WHERE id = $%d", strings.Join(sets, ", "), argN)
		res, err := r.db.ExecContext(ctx, q, args...)
		if err != nil {
			return Exercise{}, apperr.Internal("failed to update exercise")
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return Exercise{}, apperr.NotFound("exercise not found")
		}
	}
	return r.GetExercise(ctx, exerciseID)
}

func (r *PostgresRepository) DeleteExercise(ctx context.Context, exerciseID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM fitness_exercises WHERE id = $1`, exerciseID)
	if err != nil {
		return apperr.Internal("failed to delete exercise")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return apperr.NotFound("exercise not found")
	}
	return nil
}

func (r *PostgresRepository) UpsertExerciseLog(ctx context.Context, exerciseID, date string, quantity float64) (ExerciseLog, error) {
	log := ExerciseLog{ID: id.New(), ExerciseID: exerciseID, LogDate: date, Quantity: quantity}

	const q = `
		INSERT INTO fitness_exercise_logs (id, exercise_id, log_date, quantity)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (exercise_id, log_date) DO UPDATE SET quantity = EXCLUDED.quantity
		RETURNING id, exercise_id, log_date, quantity`
	err := r.db.QueryRowContext(ctx, q, log.ID, exerciseID, date, quantity).
		Scan(&log.ID, &log.ExerciseID, &log.LogDate, &log.Quantity)
	if isForeignKeyViolation(err) {
		return ExerciseLog{}, apperr.InvalidInput("exercise not found")
	}
	if err != nil {
		return ExerciseLog{}, apperr.Internal("failed to save exercise log")
	}
	return log, nil
}

func (r *PostgresRepository) ListExerciseLogs(ctx context.Context, exerciseID string) ([]ExerciseLog, error) {
	const q = `SELECT id, exercise_id, log_date, quantity FROM fitness_exercise_logs WHERE exercise_id = $1 ORDER BY log_date ASC`
	rows, err := r.db.QueryContext(ctx, q, exerciseID)
	if err != nil {
		return nil, apperr.Internal("failed to list exercise logs")
	}
	defer rows.Close()

	logs := make([]ExerciseLog, 0)
	for rows.Next() {
		var l ExerciseLog
		if err := rows.Scan(&l.ID, &l.ExerciseID, &l.LogDate, &l.Quantity); err != nil {
			return nil, apperr.Internal("failed to scan exercise log")
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list exercise logs")
	}
	return logs, nil
}

func (r *PostgresRepository) DeleteExerciseLog(ctx context.Context, logID string) error {
	return r.deleteByID(ctx, "fitness_exercise_logs", logID, "exercise log")
}

// --- weight logs ---

func (r *PostgresRepository) UpsertWeightLog(ctx context.Context, cycleID, date string, weight float64) (WeightLog, error) {
	log := WeightLog{ID: id.New(), CycleID: cycleID, LogDate: date, Weight: weight}

	const q = `
		INSERT INTO fitness_weight_logs (id, cycle_id, log_date, weight)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (cycle_id, log_date) DO UPDATE SET weight = EXCLUDED.weight
		RETURNING id, cycle_id, log_date, weight`
	err := r.db.QueryRowContext(ctx, q, log.ID, cycleID, date, weight).
		Scan(&log.ID, &log.CycleID, &log.LogDate, &log.Weight)
	if isForeignKeyViolation(err) {
		return WeightLog{}, apperr.InvalidInput("cycle not found")
	}
	if err != nil {
		return WeightLog{}, apperr.Internal("failed to save weight log")
	}
	return log, nil
}

func (r *PostgresRepository) ListWeightLogs(ctx context.Context, cycleID string) ([]WeightLog, error) {
	const q = `SELECT id, cycle_id, log_date, weight FROM fitness_weight_logs WHERE cycle_id = $1 ORDER BY log_date ASC`
	rows, err := r.db.QueryContext(ctx, q, cycleID)
	if err != nil {
		return nil, apperr.Internal("failed to list weight logs")
	}
	defer rows.Close()

	logs := make([]WeightLog, 0)
	for rows.Next() {
		var l WeightLog
		if err := rows.Scan(&l.ID, &l.CycleID, &l.LogDate, &l.Weight); err != nil {
			return nil, apperr.Internal("failed to scan weight log")
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list weight logs")
	}
	return logs, nil
}

func (r *PostgresRepository) DeleteWeightLog(ctx context.Context, logID string) error {
	return r.deleteByID(ctx, "fitness_weight_logs", logID, "weight log")
}

// --- foods ---

func (r *PostgresRepository) CreateFood(ctx context.Context, f Food) (Food, error) {
	f.ID = id.New()
	f.CreatedAt = time.Now().UTC()

	const q = `INSERT INTO fitness_foods (id, name, unit, protein_per_unit, created_at) VALUES ($1, $2, $3, $4, $5)`
	if _, err := r.db.ExecContext(ctx, q, f.ID, f.Name, f.Unit, f.ProteinPerUnit, f.CreatedAt); err != nil {
		return Food{}, apperr.Internal("failed to create food")
	}
	return f, nil
}

func (r *PostgresRepository) ListFoods(ctx context.Context) ([]Food, error) {
	const q = `SELECT id, name, unit, protein_per_unit, created_at FROM fitness_foods ORDER BY name ASC`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, apperr.Internal("failed to list foods")
	}
	defer rows.Close()

	foods := make([]Food, 0)
	for rows.Next() {
		var f Food
		if err := rows.Scan(&f.ID, &f.Name, &f.Unit, &f.ProteinPerUnit, &f.CreatedAt); err != nil {
			return nil, apperr.Internal("failed to scan food")
		}
		foods = append(foods, f)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list foods")
	}
	return foods, nil
}

func (r *PostgresRepository) GetFood(ctx context.Context, foodID string) (Food, error) {
	const q = `SELECT id, name, unit, protein_per_unit, created_at FROM fitness_foods WHERE id = $1`
	var f Food
	err := r.db.QueryRowContext(ctx, q, foodID).Scan(&f.ID, &f.Name, &f.Unit, &f.ProteinPerUnit, &f.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Food{}, apperr.NotFound("food not found")
	}
	if err != nil {
		return Food{}, apperr.Internal("failed to get food")
	}
	return f, nil
}

func (r *PostgresRepository) UpdateFood(ctx context.Context, foodID string, input UpdateFoodInput) (Food, error) {
	sets := make([]string, 0, 3)
	args := make([]any, 0, 4)
	argN := 1
	add := func(col string, val any) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, argN))
		args = append(args, val)
		argN++
	}

	if input.Name != nil {
		add("name", strings.TrimSpace(*input.Name))
	}
	if input.Unit != nil {
		add("unit", strings.TrimSpace(*input.Unit))
	}
	if input.ProteinPerUnit != nil {
		add("protein_per_unit", *input.ProteinPerUnit)
	}

	if len(sets) > 0 {
		args = append(args, foodID)
		q := fmt.Sprintf("UPDATE fitness_foods SET %s WHERE id = $%d", strings.Join(sets, ", "), argN)
		res, err := r.db.ExecContext(ctx, q, args...)
		if err != nil {
			return Food{}, apperr.Internal("failed to update food")
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return Food{}, apperr.NotFound("food not found")
		}
	}
	return r.GetFood(ctx, foodID)
}

func (r *PostgresRepository) DeleteFood(ctx context.Context, foodID string) error {
	return r.deleteByID(ctx, "fitness_foods", foodID, "food")
}

// --- protein logs ---

func (r *PostgresRepository) CreateProteinLog(ctx context.Context, l ProteinLog) (ProteinLog, error) {
	l.ID = id.New()
	l.CreatedAt = time.Now().UTC()

	const q = `INSERT INTO fitness_protein_logs (id, cycle_id, food_id, log_date, quantity, protein, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	if _, err := r.db.ExecContext(ctx, q, l.ID, l.CycleID, l.FoodID, l.LogDate, l.Quantity, l.Protein, l.CreatedAt); err != nil {
		if isForeignKeyViolation(err) {
			return ProteinLog{}, apperr.InvalidInput("cycle or food not found")
		}
		return ProteinLog{}, apperr.Internal("failed to create protein log")
	}
	return l, nil
}

func (r *PostgresRepository) ListProteinLogs(ctx context.Context, cycleID string) ([]ProteinLog, error) {
	const q = `SELECT id, cycle_id, food_id, log_date, quantity, protein, created_at
		FROM fitness_protein_logs WHERE cycle_id = $1 ORDER BY log_date DESC, created_at DESC`
	rows, err := r.db.QueryContext(ctx, q, cycleID)
	if err != nil {
		return nil, apperr.Internal("failed to list protein logs")
	}
	defer rows.Close()

	logs := make([]ProteinLog, 0)
	for rows.Next() {
		var l ProteinLog
		if err := rows.Scan(&l.ID, &l.CycleID, &l.FoodID, &l.LogDate, &l.Quantity, &l.Protein, &l.CreatedAt); err != nil {
			return nil, apperr.Internal("failed to scan protein log")
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to list protein logs")
	}
	return logs, nil
}

func (r *PostgresRepository) DeleteProteinLog(ctx context.Context, logID string) error {
	return r.deleteByID(ctx, "fitness_protein_logs", logID, "protein log")
}

func (r *PostgresRepository) ProteinDailyTotals(ctx context.Context, cycleID string) ([]ProteinDailyTotal, error) {
	const q = `SELECT log_date, SUM(protein) FROM fitness_protein_logs WHERE cycle_id = $1 GROUP BY log_date ORDER BY log_date ASC`
	rows, err := r.db.QueryContext(ctx, q, cycleID)
	if err != nil {
		return nil, apperr.Internal("failed to load protein totals")
	}
	defer rows.Close()

	totals := make([]ProteinDailyTotal, 0)
	for rows.Next() {
		var t ProteinDailyTotal
		if err := rows.Scan(&t.Date, &t.Protein); err != nil {
			return nil, apperr.Internal("failed to scan protein total")
		}
		totals = append(totals, t)
	}
	if err := rows.Err(); err != nil {
		return nil, apperr.Internal("failed to load protein totals")
	}
	return totals, nil
}

// deleteByID is the shared "DELETE FROM <table> WHERE id = $1, 404 if
// nothing matched" used by every leaf-row delete in this package.
func (r *PostgresRepository) deleteByID(ctx context.Context, table, rowID, label string) error {
	res, err := r.db.ExecContext(ctx, fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, table), rowID)
	if err != nil {
		return apperr.Internal("failed to delete " + label)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return apperr.NotFound(label + " not found")
	}
	return nil
}

// isForeignKeyViolation reports whether err is a Postgres foreign-key
// violation (SQLSTATE 23503), matching internal/upskill's helper.
func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
