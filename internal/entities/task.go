package entities

import (
	"fmt"
	"time"
)

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Priority    int64     `json:"priority"`
	Completed   bool      `json:"completed"`
	DueDate     time.Time `json:"due_date,omitempty"`
}

type TaskDTO struct {
	Title       string `form:"title" json:"title"`
	Description string `form:"description" json:"description"`
	Priority    string `form:"priority" json:"priority"`
	Completed   string `form:"completed" json:"completed"`
	DueDate     string `form:"due_date" json:"due_date,omitempty"`
}

func (t *TaskDTO) ToTask() (Task, error) {

	dueDate, err := parseDate(t.DueDate)
	if err != nil {
		return Task{}, err
	}

	return Task{
		Title:       t.Title,
		Description: t.Description,
		Priority:    parsePriority(t.Priority),
		Completed:   parseCompleted(t.Completed),
		DueDate:     dueDate,
	}, nil
}

func parsePriority(priority string) int64 {
	switch priority {
	case "low", "1":
		return 1
	case "medium", "2":
		return 2
	case "high", "3":
		return 3
	default:
		return 2
	}
}

func parseCompleted(completed string) bool {
	switch completed {
	case "true":
		return true
	case "false":
		return false
	default:
		return false
	}
}

// parseDate validates and parses a date string into time.Time
// Accepts format: YYYY-MM-DD (e.g., 2026-06-09)
func parseDate(dateStr string) (time.Time, error) {
	const dateFormat = "2006-01-02" // Go reference date: Mon Jan 2 15:04:05 MST 2006

	parsed, err := time.Parse(dateFormat, dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid due date format: '%s'. Use YYYY-MM-DD (e.g., 2026-06-09)", dateStr)
	}

	return parsed.Local(), nil
}
