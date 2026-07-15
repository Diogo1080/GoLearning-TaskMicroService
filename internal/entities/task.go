package entities

import (
	"fmt"
	"time"
)

type TaskSearch struct {
	Search     string
	Priority   string
	Completed  bool
	DueDateMin time.Time
	DueDateMax time.Time
}

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Priority    int       `json:"priority"`
	Completed   bool      `json:"completed"`
	DueDate     time.Time `json:"due_date,omitempty"`
}

type TaskDTO struct {
	ID          int    `form:"id" json:"id"`
	Title       string `form:"title" json:"title"`
	Description string `form:"description" json:"description"`
	Priority    string `form:"priority" json:"priority"`
	Completed   bool   `form:"completed" json:"completed"`
	DueDate     string `form:"due_date" json:"due_date,omitempty"`
}

func (t *TaskDTO) ToTask() (Task, error) {
	dueDate, err := parseDateTime(t.DueDate)
	if err != nil {
		return Task{}, err
	}

	return Task{
		Title:       t.Title,
		Description: t.Description,
		Priority:    parsePriorityInt(t.Priority),
		Completed:   t.Completed,
		DueDate:     dueDate,
	}, nil
}

func (t *Task) ToTaskDTO() (TaskDTO, error) {
	priority, err := parsePriorityString(t.Priority)
	if err != nil {
		return TaskDTO{}, err
	}

	return TaskDTO{
		ID:          int(t.ID),
		Title:       t.Title,
		Description: t.Description,
		Priority:    priority,
		Completed:   t.Completed,
		DueDate:     parseDateString(t.DueDate),
	}, nil
}

func parsePriorityInt(priority string) int {
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

func parsePriorityString(priority int) (string, error) {
	switch priority {
	case 1:
		return "low", nil
	case 2:
		return "medium", nil
	case 3:
		return "high", nil
	default:
		return "", fmt.Errorf("No valid priority")
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
func parseDateTime(dateStr string) (time.Time, error) {
	const dateFormat = "2006-01-02" // Go reference date: Mon Jan 2 15:04:05 MST 2006

	parsed, err := time.Parse(dateFormat, dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid due date format: '%s'. Use YYYY-MM-DD (e.g., 2026-06-09)", dateStr)
	}

	return parsed.Local(), nil
}

func parseDateString(date time.Time) string {
	return date.Format("2006-01-02")
}
