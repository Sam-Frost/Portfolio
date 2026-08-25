package settings

type DailyWorkTracker struct {
	TotalWorkHoursRequired *float64 `json:"totalWorkHoursRequired"`
}

type Settings struct {
	DailyWorkTracker DailyWorkTracker `json:"dailyWorkTracker"`
}

// DailyWorkTrackerInput mirrors DailyWorkTracker: the frontend always sends
// the whole nested object, so there's no tri-state (missing vs. explicit
// null) to track beyond the pointer already needed for "no value set".
type DailyWorkTrackerInput struct {
	TotalWorkHoursRequired *float64 `json:"totalWorkHoursRequired"`
}

// UpdateInput is a partial update: a nil DailyWorkTracker leaves that
// section unchanged, matching the label package's UpdateInput convention.
type UpdateInput struct {
	DailyWorkTracker *DailyWorkTrackerInput `json:"dailyWorkTracker"`
}
