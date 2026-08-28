export type CycleStatus = "active" | "archived";

export interface Cycle {
  id: string;
  name: string;
  startDate: string;
  weightStart: number | null;
  weightTarget: number | null;
  proteinTarget: number | null;
  status: CycleStatus;
  createdAt: string;
  archivedAt: string | null;
  exerciseCount: number;
  latestWeight: number | null;
}

export interface Exercise {
  id: string;
  cycleId: string;
  name: string;
  goalDate: string | null;
  goalQuantity: number | null;
  unit: string | null;
  createdAt: string;
  totalLogged: number;
}

export interface ExerciseLog {
  id: string;
  exerciseId: string;
  logDate: string;
  quantity: number;
}

export interface WeightLog {
  id: string;
  cycleId: string;
  logDate: string;
  weight: number;
}

// The food library is a single shared list, not cycle-scoped — it's
// managed in Settings and reused across every cycle.
export interface Food {
  id: string;
  name: string;
  unit: string;
  proteinPerUnit: number;
  createdAt: string;
}

export interface ProteinLog {
  id: string;
  cycleId: string;
  foodId: string;
  logDate: string;
  quantity: number;
  protein: number;
  createdAt: string;
}

export interface ProteinDailyTotal {
  date: string;
  protein: number;
}

// Inputs the frontend forms build.
export interface CycleInput {
  name: string;
  startDate: string;
  weightStart: number | null;
  weightTarget: number | null;
  proteinTarget: number | null;
}

export interface ExerciseInput {
  name: string;
  goalDate: string | null;
  goalQuantity: number | null;
  unit: string | null;
}
