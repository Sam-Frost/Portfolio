package fitness

import (
	"context"
	"strings"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// validateDate treats a nil or empty pointer as "not provided" (the caller
// decides whether that's allowed); a non-empty value must be YYYY-MM-DD.
func validateDate(field string, date *string) error {
	if date == nil || *date == "" {
		return nil
	}
	if _, err := time.Parse(DateLayout, *date); err != nil {
		return apperr.InvalidInput(field + " must be in YYYY-MM-DD format")
	}
	return nil
}

func requireDate(field, date string) error {
	if strings.TrimSpace(date) == "" {
		return apperr.InvalidInput(field + " is required")
	}
	if _, err := time.Parse(DateLayout, date); err != nil {
		return apperr.InvalidInput(field + " must be in YYYY-MM-DD format")
	}
	return nil
}

// requirePositive rejects a provided-but-non-positive number. A nil
// pointer is "not provided" and passes.
func requirePositive(field string, v *float64) error {
	if v == nil {
		return nil
	}
	if *v <= 0 {
		return apperr.InvalidInput(field + " must be greater than zero")
	}
	return nil
}

// --- cycles ---

func (s *Service) CreateCycle(ctx context.Context, input CreateCycleInput) (Cycle, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Cycle{}, apperr.InvalidInput("name is required")
	}
	if err := requireDate("startDate", input.StartDate); err != nil {
		return Cycle{}, err
	}
	for field, v := range map[string]*float64{
		"weightStart": input.WeightStart, "weightTarget": input.WeightTarget, "proteinTarget": input.ProteinTarget,
	} {
		if err := requirePositive(field, v); err != nil {
			return Cycle{}, err
		}
	}

	return s.repo.CreateCycle(ctx, Cycle{
		Name:          name,
		StartDate:     input.StartDate,
		WeightStart:   input.WeightStart,
		WeightTarget:  input.WeightTarget,
		ProteinTarget: input.ProteinTarget,
		Status:        StatusActive,
	})
}

func (s *Service) ListCycles(ctx context.Context) ([]Cycle, error) {
	return s.repo.ListCycles(ctx)
}

func (s *Service) GetCycle(ctx context.Context, id string) (Cycle, error) {
	return s.repo.GetCycle(ctx, id)
}

func (s *Service) ActiveCycle(ctx context.Context) (Cycle, bool, error) {
	return s.repo.GetActiveCycle(ctx)
}

func (s *Service) UpdateCycle(ctx context.Context, id string, input UpdateCycleInput) (Cycle, error) {
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return Cycle{}, apperr.InvalidInput("name is required")
	}
	if err := validateDate("startDate", input.StartDate); err != nil {
		return Cycle{}, err
	}
	for field, v := range map[string]*float64{
		"weightStart": input.WeightStart, "weightTarget": input.WeightTarget, "proteinTarget": input.ProteinTarget,
	} {
		if err := requirePositive(field, v); err != nil {
			return Cycle{}, err
		}
	}
	return s.repo.UpdateCycle(ctx, id, input)
}

func (s *Service) ArchiveCycle(ctx context.Context, id string) (Cycle, error) {
	return s.repo.ArchiveCycle(ctx, id)
}

func (s *Service) ActivateCycle(ctx context.Context, id string) (Cycle, error) {
	return s.repo.ActivateCycle(ctx, id)
}

func (s *Service) DeleteCycle(ctx context.Context, id string) error {
	return s.repo.DeleteCycle(ctx, id)
}

// --- exercises ---

func (s *Service) CreateExercise(ctx context.Context, cycleID string, input CreateExerciseInput) (Exercise, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Exercise{}, apperr.InvalidInput("name is required")
	}
	if err := validateDate("goalDate", input.GoalDate); err != nil {
		return Exercise{}, err
	}
	if err := requirePositive("goalQuantity", input.GoalQuantity); err != nil {
		return Exercise{}, err
	}

	return s.repo.CreateExercise(ctx, Exercise{
		CycleID:      cycleID,
		Name:         name,
		GoalDate:     input.GoalDate,
		GoalQuantity: input.GoalQuantity,
		Unit:         trimmedPtr(input.Unit),
	})
}

func (s *Service) ListExercises(ctx context.Context, cycleID string) ([]Exercise, error) {
	return s.repo.ListExercises(ctx, cycleID)
}

func (s *Service) GetExercise(ctx context.Context, id string) (Exercise, error) {
	return s.repo.GetExercise(ctx, id)
}

func (s *Service) UpdateExercise(ctx context.Context, id string, input UpdateExerciseInput) (Exercise, error) {
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return Exercise{}, apperr.InvalidInput("name is required")
	}
	if err := validateDate("goalDate", input.GoalDate); err != nil {
		return Exercise{}, err
	}
	if err := requirePositive("goalQuantity", input.GoalQuantity); err != nil {
		return Exercise{}, err
	}
	return s.repo.UpdateExercise(ctx, id, input)
}

func (s *Service) DeleteExercise(ctx context.Context, id string) error {
	return s.repo.DeleteExercise(ctx, id)
}

func (s *Service) UpsertExerciseLog(ctx context.Context, exerciseID string, input UpsertExerciseLogInput) (ExerciseLog, error) {
	if err := requireDate("date", input.Date); err != nil {
		return ExerciseLog{}, err
	}
	if input.Quantity <= 0 {
		return ExerciseLog{}, apperr.InvalidInput("quantity must be greater than zero")
	}
	return s.repo.UpsertExerciseLog(ctx, exerciseID, input.Date, input.Quantity)
}

func (s *Service) ListExerciseLogs(ctx context.Context, exerciseID string) ([]ExerciseLog, error) {
	return s.repo.ListExerciseLogs(ctx, exerciseID)
}

func (s *Service) DeleteExerciseLog(ctx context.Context, id string) error {
	return s.repo.DeleteExerciseLog(ctx, id)
}

// --- weight logs ---

func (s *Service) UpsertWeightLog(ctx context.Context, cycleID string, input UpsertWeightLogInput) (WeightLog, error) {
	if err := requireDate("date", input.Date); err != nil {
		return WeightLog{}, err
	}
	if input.Weight <= 0 {
		return WeightLog{}, apperr.InvalidInput("weight must be greater than zero")
	}
	return s.repo.UpsertWeightLog(ctx, cycleID, input.Date, input.Weight)
}

func (s *Service) ListWeightLogs(ctx context.Context, cycleID string) ([]WeightLog, error) {
	return s.repo.ListWeightLogs(ctx, cycleID)
}

func (s *Service) DeleteWeightLog(ctx context.Context, id string) error {
	return s.repo.DeleteWeightLog(ctx, id)
}

// --- foods ---

func (s *Service) CreateFood(ctx context.Context, input CreateFoodInput) (Food, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Food{}, apperr.InvalidInput("name is required")
	}
	unit := strings.TrimSpace(input.Unit)
	if unit == "" {
		return Food{}, apperr.InvalidInput("unit is required")
	}
	if input.ProteinPerUnit <= 0 {
		return Food{}, apperr.InvalidInput("proteinPerUnit must be greater than zero")
	}

	return s.repo.CreateFood(ctx, Food{
		Name:           name,
		Unit:           unit,
		ProteinPerUnit: input.ProteinPerUnit,
	})
}

func (s *Service) ListFoods(ctx context.Context) ([]Food, error) {
	return s.repo.ListFoods(ctx)
}

func (s *Service) UpdateFood(ctx context.Context, id string, input UpdateFoodInput) (Food, error) {
	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		return Food{}, apperr.InvalidInput("name is required")
	}
	if input.Unit != nil && strings.TrimSpace(*input.Unit) == "" {
		return Food{}, apperr.InvalidInput("unit is required")
	}
	if input.ProteinPerUnit != nil && *input.ProteinPerUnit <= 0 {
		return Food{}, apperr.InvalidInput("proteinPerUnit must be greater than zero")
	}
	return s.repo.UpdateFood(ctx, id, input)
}

func (s *Service) DeleteFood(ctx context.Context, id string) error {
	return s.repo.DeleteFood(ctx, id)
}

// --- protein logs ---

// CreateProteinLog looks up the food to snapshot protein = quantity *
// proteinPerUnit onto the row, so a later edit to the food's protein
// content doesn't retroactively change past intake.
func (s *Service) CreateProteinLog(ctx context.Context, cycleID string, input CreateProteinLogInput) (ProteinLog, error) {
	if strings.TrimSpace(input.FoodID) == "" {
		return ProteinLog{}, apperr.InvalidInput("foodId is required")
	}
	if err := requireDate("date", input.Date); err != nil {
		return ProteinLog{}, err
	}
	if input.Quantity <= 0 {
		return ProteinLog{}, apperr.InvalidInput("quantity must be greater than zero")
	}

	food, err := s.repo.GetFood(ctx, input.FoodID)
	if err != nil {
		return ProteinLog{}, err
	}

	return s.repo.CreateProteinLog(ctx, ProteinLog{
		CycleID:  cycleID,
		FoodID:   input.FoodID,
		LogDate:  input.Date,
		Quantity: input.Quantity,
		Protein:  input.Quantity * food.ProteinPerUnit,
	})
}

func (s *Service) ListProteinLogs(ctx context.Context, cycleID string) ([]ProteinLog, error) {
	return s.repo.ListProteinLogs(ctx, cycleID)
}

func (s *Service) DeleteProteinLog(ctx context.Context, id string) error {
	return s.repo.DeleteProteinLog(ctx, id)
}

func (s *Service) ProteinDailyTotals(ctx context.Context, cycleID string) ([]ProteinDailyTotal, error) {
	return s.repo.ProteinDailyTotals(ctx, cycleID)
}

func trimmedPtr(s *string) *string {
	if s == nil {
		return nil
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return nil
	}
	return &t
}
