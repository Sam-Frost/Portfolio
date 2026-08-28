package fitness

import (
	"net/http"

	"github.com/Sam-Frost/portfolio/internal/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/fitness/cycles", h.listCycles)
	mux.HandleFunc("POST /api/fitness/cycles", h.createCycle)
	mux.HandleFunc("GET /api/fitness/cycles/active", h.activeCycle)
	mux.HandleFunc("GET /api/fitness/cycles/{id}", h.getCycle)
	mux.HandleFunc("PATCH /api/fitness/cycles/{id}", h.updateCycle)
	mux.HandleFunc("DELETE /api/fitness/cycles/{id}", h.deleteCycle)
	mux.HandleFunc("POST /api/fitness/cycles/{id}/archive", h.archiveCycle)
	mux.HandleFunc("POST /api/fitness/cycles/{id}/activate", h.activateCycle)

	mux.HandleFunc("GET /api/fitness/cycles/{id}/exercises", h.listExercises)
	mux.HandleFunc("POST /api/fitness/cycles/{id}/exercises", h.createExercise)
	mux.HandleFunc("GET /api/fitness/exercises/{id}", h.getExercise)
	mux.HandleFunc("PATCH /api/fitness/exercises/{id}", h.updateExercise)
	mux.HandleFunc("DELETE /api/fitness/exercises/{id}", h.deleteExercise)
	mux.HandleFunc("GET /api/fitness/exercises/{id}/logs", h.listExerciseLogs)
	// POST upserts the entry for the date in the body (at most one log per
	// exercise per day) — see UpsertExerciseLog.
	mux.HandleFunc("POST /api/fitness/exercises/{id}/logs", h.upsertExerciseLog)
	mux.HandleFunc("DELETE /api/fitness/exercise-logs/{id}", h.deleteExerciseLog)

	mux.HandleFunc("GET /api/fitness/cycles/{id}/weight-logs", h.listWeightLogs)
	mux.HandleFunc("POST /api/fitness/cycles/{id}/weight-logs", h.upsertWeightLog)
	mux.HandleFunc("DELETE /api/fitness/weight-logs/{id}", h.deleteWeightLog)

	// The food library is a single shared list, not cycle-scoped.
	mux.HandleFunc("GET /api/fitness/foods", h.listFoods)
	mux.HandleFunc("POST /api/fitness/foods", h.createFood)
	mux.HandleFunc("PATCH /api/fitness/foods/{id}", h.updateFood)
	mux.HandleFunc("DELETE /api/fitness/foods/{id}", h.deleteFood)

	mux.HandleFunc("GET /api/fitness/cycles/{id}/protein-logs", h.listProteinLogs)
	mux.HandleFunc("POST /api/fitness/cycles/{id}/protein-logs", h.createProteinLog)
	mux.HandleFunc("DELETE /api/fitness/protein-logs/{id}", h.deleteProteinLog)
	mux.HandleFunc("GET /api/fitness/cycles/{id}/protein-daily", h.proteinDaily)
}

// --- cycles ---

func (h *Handler) listCycles(w http.ResponseWriter, r *http.Request) {
	cycles, err := h.service.ListCycles(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, cycles)
}

func (h *Handler) createCycle(w http.ResponseWriter, r *http.Request) {
	var input CreateCycleInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	cycle, err := h.service.CreateCycle(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, cycle)
}

// activeCycle returns 204 (not a JSON null) when no cycle is active,
// matching /api/work-sessions/current.
func (h *Handler) activeCycle(w http.ResponseWriter, r *http.Request) {
	cycle, ok, err := h.service.ActiveCycle(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, cycle)
}

func (h *Handler) getCycle(w http.ResponseWriter, r *http.Request) {
	cycle, err := h.service.GetCycle(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, cycle)
}

func (h *Handler) updateCycle(w http.ResponseWriter, r *http.Request) {
	var input UpdateCycleInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	cycle, err := h.service.UpdateCycle(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, cycle)
}

func (h *Handler) archiveCycle(w http.ResponseWriter, r *http.Request) {
	cycle, err := h.service.ArchiveCycle(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, cycle)
}

func (h *Handler) activateCycle(w http.ResponseWriter, r *http.Request) {
	cycle, err := h.service.ActivateCycle(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, cycle)
}

func (h *Handler) deleteCycle(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteCycle(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- exercises ---

func (h *Handler) listExercises(w http.ResponseWriter, r *http.Request) {
	exercises, err := h.service.ListExercises(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, exercises)
}

func (h *Handler) createExercise(w http.ResponseWriter, r *http.Request) {
	var input CreateExerciseInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	exercise, err := h.service.CreateExercise(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, exercise)
}

func (h *Handler) getExercise(w http.ResponseWriter, r *http.Request) {
	exercise, err := h.service.GetExercise(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, exercise)
}

func (h *Handler) updateExercise(w http.ResponseWriter, r *http.Request) {
	var input UpdateExerciseInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	exercise, err := h.service.UpdateExercise(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, exercise)
}

func (h *Handler) deleteExercise(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteExercise(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listExerciseLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.service.ListExerciseLogs(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, logs)
}

func (h *Handler) upsertExerciseLog(w http.ResponseWriter, r *http.Request) {
	var input UpsertExerciseLogInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	log, err := h.service.UpsertExerciseLog(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, log)
}

func (h *Handler) deleteExerciseLog(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteExerciseLog(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- weight logs ---

func (h *Handler) listWeightLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.service.ListWeightLogs(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, logs)
}

func (h *Handler) upsertWeightLog(w http.ResponseWriter, r *http.Request) {
	var input UpsertWeightLogInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	log, err := h.service.UpsertWeightLog(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, log)
}

func (h *Handler) deleteWeightLog(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteWeightLog(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- foods ---

func (h *Handler) listFoods(w http.ResponseWriter, r *http.Request) {
	foods, err := h.service.ListFoods(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, foods)
}

func (h *Handler) createFood(w http.ResponseWriter, r *http.Request) {
	var input CreateFoodInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	food, err := h.service.CreateFood(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, food)
}

func (h *Handler) updateFood(w http.ResponseWriter, r *http.Request) {
	var input UpdateFoodInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	food, err := h.service.UpdateFood(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, food)
}

func (h *Handler) deleteFood(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteFood(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- protein logs ---

func (h *Handler) listProteinLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.service.ListProteinLogs(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, logs)
}

func (h *Handler) createProteinLog(w http.ResponseWriter, r *http.Request) {
	var input CreateProteinLogInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, err)
		return
	}
	log, err := h.service.CreateProteinLog(r.Context(), r.PathValue("id"), input)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, log)
}

func (h *Handler) deleteProteinLog(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteProteinLog(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) proteinDaily(w http.ResponseWriter, r *http.Request) {
	totals, err := h.service.ProteinDailyTotals(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, totals)
}
