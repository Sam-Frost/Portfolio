package scheduler

import (
	"fmt"
	"strings"
	"time"

	"github.com/Sam-Frost/portfolio/internal/notification"
	"github.com/Sam-Frost/portfolio/internal/todo"
)

// digestDue reports the current IST calendar date ("YYYY-MM-DD") and
// whether the local wall-clock time is at or past morningTime ("HH:MM")
// today. The caller still checks the dedup ledger — this only answers "are
// we past the trigger time for today".
func digestDue(now time.Time, morningTime string, loc *time.Location) (istDate string, due bool) {
	local := now.In(loc)
	istDate = local.Format(todo.TargetDateLayout)

	hh, mm, ok := parseHHMM(morningTime)
	if !ok {
		return istDate, false
	}
	trigger := time.Date(local.Year(), local.Month(), local.Day(), hh, mm, 0, 0, loc)
	return istDate, !local.Before(trigger)
}

func parseHHMM(s string) (hh, mm int, ok bool) {
	if _, err := fmt.Sscanf(s, "%d:%d", &hh, &mm); err != nil {
		return 0, 0, false
	}
	if hh < 0 || hh > 23 || mm < 0 || mm > 59 {
		return 0, 0, false
	}
	return hh, mm, true
}

// overdueTodos keeps the not-done todos whose target date is today or
// earlier — "due today or past due", matching the user's ask.
func overdueTodos(todos []todo.Todo, istToday string) []todo.Todo {
	out := make([]todo.Todo, 0)
	for _, t := range todos {
		if t.Done || t.TargetDate == nil {
			continue
		}
		if *t.TargetDate <= istToday { // ISO dates compare lexicographically
			out = append(out, t)
		}
	}
	return out
}

func buildDigestMessage(overdue []todo.Todo, istToday string) notification.Message {
	var lines []string
	for _, t := range overdue {
		when := *t.TargetDate
		if when == istToday {
			lines = append(lines, "• "+t.Name+" (due today)")
		} else {
			lines = append(lines, "• "+t.Name+" (due "+when+")")
		}
	}

	noun := "todos"
	if len(overdue) == 1 {
		noun = "todo"
	}
	return notification.Message{
		Title: fmt.Sprintf("%d %s due or overdue", len(overdue), noun),
		Body:  strings.Join(lines, "\n"),
		URL:   "/todos",
		Tag:   "morning-digest",
	}
}
