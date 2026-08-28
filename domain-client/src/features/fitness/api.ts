import { apiRequest } from "../../lib/apiClient";
import type {
  Cycle,
  CycleInput,
  Exercise,
  ExerciseInput,
  ExerciseLog,
  Food,
  ProteinDailyTotal,
  ProteinLog,
  WeightLog,
} from "./types";

// --- cycles ---

export function fetchCycles(): Promise<Cycle[]> {
  return apiRequest<Cycle[]>("/api/fitness/cycles");
}

export function fetchCycle(id: string): Promise<Cycle> {
  return apiRequest<Cycle>(`/api/fitness/cycles/${id}`);
}

// undefined on 204 — no active cycle.
export function fetchActiveCycle(): Promise<Cycle | undefined> {
  return apiRequest<Cycle | undefined>("/api/fitness/cycles/active");
}

export function createCycle(input: CycleInput): Promise<Cycle> {
  return apiRequest<Cycle>("/api/fitness/cycles", { method: "POST", body: JSON.stringify(input) });
}

// The edit dialog always submits the full current set; the backend treats a
// nil field as "leave unchanged", so every field is sent explicitly.
export function updateCycle(id: string, input: CycleInput): Promise<Cycle> {
  return apiRequest<Cycle>(`/api/fitness/cycles/${id}`, { method: "PATCH", body: JSON.stringify(input) });
}

export function archiveCycle(id: string): Promise<Cycle> {
  return apiRequest<Cycle>(`/api/fitness/cycles/${id}/archive`, { method: "POST" });
}

export function activateCycle(id: string): Promise<Cycle> {
  return apiRequest<Cycle>(`/api/fitness/cycles/${id}/activate`, { method: "POST" });
}

export function deleteCycle(id: string): Promise<void> {
  return apiRequest<void>(`/api/fitness/cycles/${id}`, { method: "DELETE" });
}

// --- exercises ---

export function fetchExercises(cycleId: string): Promise<Exercise[]> {
  return apiRequest<Exercise[]>(`/api/fitness/cycles/${cycleId}/exercises`);
}

export function fetchExercise(id: string): Promise<Exercise> {
  return apiRequest<Exercise>(`/api/fitness/exercises/${id}`);
}

export function createExercise(cycleId: string, input: ExerciseInput): Promise<Exercise> {
  return apiRequest<Exercise>(`/api/fitness/cycles/${cycleId}/exercises`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateExercise(id: string, input: ExerciseInput): Promise<Exercise> {
  return apiRequest<Exercise>(`/api/fitness/exercises/${id}`, {
    method: "PATCH",
    body: JSON.stringify({ ...input, goalDate: input.goalDate ?? "", unit: input.unit ?? "" }),
  });
}

export function deleteExercise(id: string): Promise<void> {
  return apiRequest<void>(`/api/fitness/exercises/${id}`, { method: "DELETE" });
}

export function fetchExerciseLogs(exerciseId: string): Promise<ExerciseLog[]> {
  return apiRequest<ExerciseLog[]>(`/api/fitness/exercises/${exerciseId}/logs`);
}

// POST upserts the entry for `date` (at most one log per exercise per day).
export function upsertExerciseLog(exerciseId: string, date: string, quantity: number): Promise<ExerciseLog> {
  return apiRequest<ExerciseLog>(`/api/fitness/exercises/${exerciseId}/logs`, {
    method: "POST",
    body: JSON.stringify({ date, quantity }),
  });
}

export function deleteExerciseLog(id: string): Promise<void> {
  return apiRequest<void>(`/api/fitness/exercise-logs/${id}`, { method: "DELETE" });
}

// --- weight logs ---

export function fetchWeightLogs(cycleId: string): Promise<WeightLog[]> {
  return apiRequest<WeightLog[]>(`/api/fitness/cycles/${cycleId}/weight-logs`);
}

export function upsertWeightLog(cycleId: string, date: string, weight: number): Promise<WeightLog> {
  return apiRequest<WeightLog>(`/api/fitness/cycles/${cycleId}/weight-logs`, {
    method: "POST",
    body: JSON.stringify({ date, weight }),
  });
}

export function deleteWeightLog(id: string): Promise<void> {
  return apiRequest<void>(`/api/fitness/weight-logs/${id}`, { method: "DELETE" });
}

// --- foods & protein logs ---

// The food library is shared across all cycles — not cycle-scoped.
export function fetchFoods(): Promise<Food[]> {
  return apiRequest<Food[]>("/api/fitness/foods");
}

export function createFood(input: {
  name: string;
  unit: string;
  proteinPerUnit: number;
}): Promise<Food> {
  return apiRequest<Food>("/api/fitness/foods", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateFood(
  id: string,
  input: { name: string; unit: string; proteinPerUnit: number },
): Promise<Food> {
  return apiRequest<Food>(`/api/fitness/foods/${id}`, { method: "PATCH", body: JSON.stringify(input) });
}

export function deleteFood(id: string): Promise<void> {
  return apiRequest<void>(`/api/fitness/foods/${id}`, { method: "DELETE" });
}

export function fetchProteinLogs(cycleId: string): Promise<ProteinLog[]> {
  return apiRequest<ProteinLog[]>(`/api/fitness/cycles/${cycleId}/protein-logs`);
}

export function createProteinLog(
  cycleId: string,
  input: { foodId: string; date: string; quantity: number },
): Promise<ProteinLog> {
  return apiRequest<ProteinLog>(`/api/fitness/cycles/${cycleId}/protein-logs`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function deleteProteinLog(id: string): Promise<void> {
  return apiRequest<void>(`/api/fitness/protein-logs/${id}`, { method: "DELETE" });
}

export function fetchProteinDaily(cycleId: string): Promise<ProteinDailyTotal[]> {
  return apiRequest<ProteinDailyTotal[]>(`/api/fitness/cycles/${cycleId}/protein-daily`);
}
