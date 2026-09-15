package domain

import (
	"fmt"
	"time"
)

type TaskSearch struct {
	UserID     int
	Search     string
	Priority   string
	Completed  string
	DueDateMin string
	DueDateMax string
	OrderBy    string
	Limit      int
}

type Task struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"-"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Priority    int       `json:"priority"`
	Completed   bool      `json:"completed"`
	DueDate     time.Time `json:"dueDate"`
}

type TaskDTO struct {
	ID          int    `form:"id" json:"id"`
	Title       string `form:"title" json:"title"`
	Description string `form:"description" json:"description"`
	Priority    int    `form:"priority" json:"priority"`
	Completed   bool   `form:"completed" json:"completed"`
	DueDate     string `form:"dueDate" json:"dueDate,omitempty"`
}

func (t *TaskSearch) Sanatise() error {
	return nil
}

func (t *TaskDTO) ToTask() (Task, error) {
	dueDate, err := parseDateTime(t.DueDate)
	if err != nil {
		return Task{}, err
	}

	return Task{
		Title:       t.Title,
		Description: t.Description,
		Priority:    t.Priority,
		Completed:   t.Completed,
		DueDate:     dueDate,
	}, nil
}

func (t *Task) ToTaskDTO() TaskDTO {
	return TaskDTO{
		ID:          int(t.ID),
		Title:       t.Title,
		Description: t.Description,
		Priority:    t.Priority,
		Completed:   t.Completed,
		DueDate:     parseDateString(t.DueDate),
	}
}

func parsePriorityInt(priority string) int {
	switch priority {
	case "low", "1":
		return 1
	case "medium", "2":
		return 2
	case "high", "3":
		return 3
	}
	return 0
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
	return date.String()
}
