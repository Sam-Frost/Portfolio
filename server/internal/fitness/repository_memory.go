package fitness

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
	"github.com/Sam-Frost/portfolio/internal/id"
)

// MemoryRepository is an in-memory Repository for tests. It mirrors
// PostgresRepository's observable behaviour (one active cycle, per-day
// upsert, computed TotalLogged, cascade delete) but not its storage.
type MemoryRepository struct {
	mu           sync.Mutex
	cycles       map[string]Cycle
	exercises    map[string]Exercise
	exerciseLogs map[string]ExerciseLog
	weightLogs   map[string]WeightLog
	foods        map[string]Food
	proteinLogs  map[string]ProteinLog
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		cycles:       make(map[string]Cycle),
		exercises:    make(map[string]Exercise),
		exerciseLogs: make(map[string]ExerciseLog),
		weightLogs:   make(map[string]WeightLog),
		foods:        make(map[string]Food),
		proteinLogs:  make(map[string]ProteinLog),
	}
}

// --- cycles ---

func (r *MemoryRepository) CreateCycle(_ context.Context, c Cycle) (Cycle, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	for cid, existing := range r.cycles {
		if existing.Status == StatusActive {
			existing.Status = StatusArchived
			existing.ArchivedAt = &now
			r.cycles[cid] = existing
		}
	}

	c.ID = id.New()
	c.CreatedAt = now
	c.Status = StatusActive
	c.ArchivedAt = nil
	r.cycles[c.ID] = c
	return c, nil
}

// cycleWithStatsLocked stamps the computed ExerciseCount / LatestWeight
// fields, mirroring PostgresRepository's cycleSelect. Call with r.mu held.
func (r *MemoryRepository) cycleWithStatsLocked(c Cycle) Cycle {
	for _, e := range r.exercises {
		if e.CycleID == c.ID {
			c.ExerciseCount++
		}
	}
	var latest *WeightLog
	for i, l := range r.weightLogs {
		if l.CycleID == c.ID && (latest == nil || l.LogDate > latest.LogDate) {
			w := r.weightLogs[i]
			latest = &w
		}
	}
	if latest != nil {
		v := latest.Weight
		c.LatestWeight = &v
	}
	return c
}

func (r *MemoryRepository) ListCycles(_ context.Context) ([]Cycle, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	cycles := make([]Cycle, 0, len(r.cycles))
	for _, c := range r.cycles {
		cycles = append(cycles, r.cycleWithStatsLocked(c))
	}
	sort.Slice(cycles, func(i, j int) bool {
		if (cycles[i].Status == StatusActive) != (cycles[j].Status == StatusActive) {
			return cycles[i].Status == StatusActive
		}
		return cycles[i].CreatedAt.After(cycles[j].CreatedAt)
	})
	return cycles, nil
}

func (r *MemoryRepository) GetCycle(_ context.Context, cycleID string) (Cycle, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.cycles[cycleID]
	if !ok {
		return Cycle{}, apperr.NotFound("cycle not found")
	}
	return r.cycleWithStatsLocked(c), nil
}

func (r *MemoryRepository) GetActiveCycle(_ context.Context) (Cycle, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, c := range r.cycles {
		if c.Status == StatusActive {
			return r.cycleWithStatsLocked(c), true, nil
		}
	}
	return Cycle{}, false, nil
}

func (r *MemoryRepository) UpdateCycle(_ context.Context, cycleID string, input UpdateCycleInput) (Cycle, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.cycles[cycleID]
	if !ok {
		return Cycle{}, apperr.NotFound("cycle not found")
	}
	if input.Name != nil {
		c.Name = strings.TrimSpace(*input.Name)
	}
	if input.StartDate != nil {
		c.StartDate = *input.StartDate
	}
	if input.WeightStart != nil {
		c.WeightStart = input.WeightStart
	}
	if input.WeightTarget != nil {
		c.WeightTarget = input.WeightTarget
	}
	if input.ProteinTarget != nil {
		c.ProteinTarget = input.ProteinTarget
	}
	r.cycles[cycleID] = c
	return c, nil
}

func (r *MemoryRepository) ArchiveCycle(_ context.Context, cycleID string) (Cycle, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.cycles[cycleID]
	if !ok {
		return Cycle{}, apperr.NotFound("cycle not found")
	}
	if c.Status == StatusActive {
		now := time.Now().UTC()
		c.Status = StatusArchived
		c.ArchivedAt = &now
		r.cycles[cycleID] = c
	}
	return c, nil
}

func (r *MemoryRepository) ActivateCycle(_ context.Context, cycleID string) (Cycle, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.cycles[cycleID]
	if !ok {
		return Cycle{}, apperr.NotFound("cycle not found")
	}

	now := time.Now().UTC()
	for cid, existing := range r.cycles {
		if cid != cycleID && existing.Status == StatusActive {
			existing.Status = StatusArchived
			existing.ArchivedAt = &now
			r.cycles[cid] = existing
		}
	}

	c.Status = StatusActive
	c.ArchivedAt = nil
	r.cycles[cycleID] = c
	return r.cycleWithStatsLocked(c), nil
}

func (r *MemoryRepository) DeleteCycle(_ context.Context, cycleID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.cycles[cycleID]; !ok {
		return apperr.NotFound("cycle not found")
	}
	delete(r.cycles, cycleID)

	for eid, e := range r.exercises {
		if e.CycleID == cycleID {
			delete(r.exercises, eid)
			for lid, l := range r.exerciseLogs {
				if l.ExerciseID == eid {
					delete(r.exerciseLogs, lid)
				}
			}
		}
	}
	for lid, l := range r.weightLogs {
		if l.CycleID == cycleID {
			delete(r.weightLogs, lid)
		}
	}
	// Foods are a shared library, not owned by the cycle — they survive.
	for lid, l := range r.proteinLogs {
		if l.CycleID == cycleID {
			delete(r.proteinLogs, lid)
		}
	}
	return nil
}

// --- exercises ---

func (r *MemoryRepository) CreateExercise(_ context.Context, e Exercise) (Exercise, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.cycles[e.CycleID]; !ok {
		return Exercise{}, apperr.InvalidInput("cycle not found")
	}
	e.ID = id.New()
	e.CreatedAt = time.Now().UTC()
	e.TotalLogged = 0
	r.exercises[e.ID] = e
	return e, nil
}

// totalLoggedLocked must be called with r.mu held.
func (r *MemoryRepository) totalLoggedLocked(exerciseID string) float64 {
	var total float64
	for _, l := range r.exerciseLogs {
		if l.ExerciseID == exerciseID {
			total += l.Quantity
		}
	}
	return total
}

func (r *MemoryRepository) ListExercises(_ context.Context, cycleID string) ([]Exercise, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	exercises := make([]Exercise, 0)
	for _, e := range r.exercises {
		if e.CycleID != cycleID {
			continue
		}
		e.TotalLogged = r.totalLoggedLocked(e.ID)
		exercises = append(exercises, e)
	}
	sort.Slice(exercises, func(i, j int) bool {
		return exercises[i].CreatedAt.After(exercises[j].CreatedAt)
	})
	return exercises, nil
}

func (r *MemoryRepository) GetExercise(_ context.Context, exerciseID string) (Exercise, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	e, ok := r.exercises[exerciseID]
	if !ok {
		return Exercise{}, apperr.NotFound("exercise not found")
	}
	e.TotalLogged = r.totalLoggedLocked(e.ID)
	return e, nil
}

func (r *MemoryRepository) UpdateExercise(_ context.Context, exerciseID string, input UpdateExerciseInput) (Exercise, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	e, ok := r.exercises[exerciseID]
	if !ok {
		return Exercise{}, apperr.NotFound("exercise not found")
	}
	if input.Name != nil {
		e.Name = strings.TrimSpace(*input.Name)
	}
	if input.GoalDate != nil {
		if *input.GoalDate == "" {
			e.GoalDate = nil
		} else {
			e.GoalDate = input.GoalDate
		}
	}
	if input.GoalQuantity != nil {
		e.GoalQuantity = input.GoalQuantity
	}
	if input.Unit != nil {
		if trimmed := strings.TrimSpace(*input.Unit); trimmed == "" {
			e.Unit = nil
		} else {
			e.Unit = &trimmed
		}
	}
	r.exercises[exerciseID] = e
	e.TotalLogged = r.totalLoggedLocked(e.ID)
	return e, nil
}

func (r *MemoryRepository) DeleteExercise(_ context.Context, exerciseID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.exercises[exerciseID]; !ok {
		return apperr.NotFound("exercise not found")
	}
	delete(r.exercises, exerciseID)
	for lid, l := range r.exerciseLogs {
		if l.ExerciseID == exerciseID {
			delete(r.exerciseLogs, lid)
		}
	}
	return nil
}

func (r *MemoryRepository) UpsertExerciseLog(_ context.Context, exerciseID, date string, quantity float64) (ExerciseLog, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.exercises[exerciseID]; !ok {
		return ExerciseLog{}, apperr.InvalidInput("exercise not found")
	}
	for lid, l := range r.exerciseLogs {
		if l.ExerciseID == exerciseID && l.LogDate == date {
			l.Quantity = quantity
			r.exerciseLogs[lid] = l
			return l, nil
		}
	}
	log := ExerciseLog{ID: id.New(), ExerciseID: exerciseID, LogDate: date, Quantity: quantity}
	r.exerciseLogs[log.ID] = log
	return log, nil
}

func (r *MemoryRepository) ListExerciseLogs(_ context.Context, exerciseID string) ([]ExerciseLog, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	logs := make([]ExerciseLog, 0)
	for _, l := range r.exerciseLogs {
		if l.ExerciseID == exerciseID {
			logs = append(logs, l)
		}
	}
	sort.Slice(logs, func(i, j int) bool { return logs[i].LogDate < logs[j].LogDate })
	return logs, nil
}

func (r *MemoryRepository) DeleteExerciseLog(_ context.Context, logID string) error {
	return r.deleteFrom(logID, func(k string) bool { _, ok := r.exerciseLogs[k]; return ok }, func(k string) { delete(r.exerciseLogs, k) }, "exercise log")
}

// --- weight logs ---

func (r *MemoryRepository) UpsertWeightLog(_ context.Context, cycleID, date string, weight float64) (WeightLog, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.cycles[cycleID]; !ok {
		return WeightLog{}, apperr.InvalidInput("cycle not found")
	}
	for lid, l := range r.weightLogs {
		if l.CycleID == cycleID && l.LogDate == date {
			l.Weight = weight
			r.weightLogs[lid] = l
			return l, nil
		}
	}
	log := WeightLog{ID: id.New(), CycleID: cycleID, LogDate: date, Weight: weight}
	r.weightLogs[log.ID] = log
	return log, nil
}

func (r *MemoryRepository) ListWeightLogs(_ context.Context, cycleID string) ([]WeightLog, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	logs := make([]WeightLog, 0)
	for _, l := range r.weightLogs {
		if l.CycleID == cycleID {
			logs = append(logs, l)
		}
	}
	sort.Slice(logs, func(i, j int) bool { return logs[i].LogDate < logs[j].LogDate })
	return logs, nil
}

func (r *MemoryRepository) DeleteWeightLog(_ context.Context, logID string) error {
	return r.deleteFrom(logID, func(k string) bool { _, ok := r.weightLogs[k]; return ok }, func(k string) { delete(r.weightLogs, k) }, "weight log")
}

// --- foods ---

func (r *MemoryRepository) CreateFood(_ context.Context, f Food) (Food, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	f.ID = id.New()
	f.CreatedAt = time.Now().UTC()
	r.foods[f.ID] = f
	return f, nil
}

func (r *MemoryRepository) ListFoods(_ context.Context) ([]Food, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	foods := make([]Food, 0, len(r.foods))
	for _, f := range r.foods {
		foods = append(foods, f)
	}
	sort.Slice(foods, func(i, j int) bool { return strings.ToLower(foods[i].Name) < strings.ToLower(foods[j].Name) })
	return foods, nil
}

func (r *MemoryRepository) GetFood(_ context.Context, foodID string) (Food, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	f, ok := r.foods[foodID]
	if !ok {
		return Food{}, apperr.NotFound("food not found")
	}
	return f, nil
}

func (r *MemoryRepository) UpdateFood(_ context.Context, foodID string, input UpdateFoodInput) (Food, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	f, ok := r.foods[foodID]
	if !ok {
		return Food{}, apperr.NotFound("food not found")
	}
	if input.Name != nil {
		f.Name = strings.TrimSpace(*input.Name)
	}
	if input.Unit != nil {
		f.Unit = strings.TrimSpace(*input.Unit)
	}
	if input.ProteinPerUnit != nil {
		f.ProteinPerUnit = *input.ProteinPerUnit
	}
	r.foods[foodID] = f
	return f, nil
}

func (r *MemoryRepository) DeleteFood(_ context.Context, foodID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.foods[foodID]; !ok {
		return apperr.NotFound("food not found")
	}
	delete(r.foods, foodID)
	for lid, l := range r.proteinLogs {
		if l.FoodID == foodID {
			delete(r.proteinLogs, lid)
		}
	}
	return nil
}

// --- protein logs ---

func (r *MemoryRepository) CreateProteinLog(_ context.Context, l ProteinLog) (ProteinLog, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.cycles[l.CycleID]; !ok {
		return ProteinLog{}, apperr.InvalidInput("cycle not found")
	}
	if _, ok := r.foods[l.FoodID]; !ok {
		return ProteinLog{}, apperr.InvalidInput("food not found")
	}
	l.ID = id.New()
	l.CreatedAt = time.Now().UTC()
	r.proteinLogs[l.ID] = l
	return l, nil
}

func (r *MemoryRepository) ListProteinLogs(_ context.Context, cycleID string) ([]ProteinLog, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	logs := make([]ProteinLog, 0)
	for _, l := range r.proteinLogs {
		if l.CycleID == cycleID {
			logs = append(logs, l)
		}
	}
	sort.Slice(logs, func(i, j int) bool {
		if logs[i].LogDate != logs[j].LogDate {
			return logs[i].LogDate > logs[j].LogDate
		}
		return logs[i].CreatedAt.After(logs[j].CreatedAt)
	})
	return logs, nil
}

func (r *MemoryRepository) DeleteProteinLog(_ context.Context, logID string) error {
	return r.deleteFrom(logID, func(k string) bool { _, ok := r.proteinLogs[k]; return ok }, func(k string) { delete(r.proteinLogs, k) }, "protein log")
}

func (r *MemoryRepository) ProteinDailyTotals(_ context.Context, cycleID string) ([]ProteinDailyTotal, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	byDate := make(map[string]float64)
	for _, l := range r.proteinLogs {
		if l.CycleID == cycleID {
			byDate[l.LogDate] += l.Protein
		}
	}
	dates := make([]string, 0, len(byDate))
	for d := range byDate {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	totals := make([]ProteinDailyTotal, 0, len(dates))
	for _, d := range dates {
		totals = append(totals, ProteinDailyTotal{Date: d, Protein: byDate[d]})
	}
	return totals, nil
}

// deleteFrom is the locked "exists? delete : NotFound" shared by the
// leaf-row deletes.
func (r *MemoryRepository) deleteFrom(key string, exists func(string) bool, del func(string), label string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !exists(key) {
		return apperr.NotFound(label + " not found")
	}
	del(key)
	return nil
}
