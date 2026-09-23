package domain

import (
	"fmt"
	"strconv"
	"strings"
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
	if t.UserID <= 0 {
		return ErrBadRequest
	}

	t.Search = strings.TrimSpace(t.Search)
	if t.Limit <= 0 {
		t.Limit = 100
	}
	if t.Limit > 100 {
		t.Limit = 100
	}

	if t.Priority != "" {
		priority, err := strconv.Atoi(t.Priority)
		if err != nil || priority < 1 || priority > 3 {
			return ErrBadRequest
		}
		t.Priority = strconv.Itoa(priority)
	}

	if t.Completed != "" {
		completed, err := strconv.ParseBool(t.Completed)
		if err != nil {
			return ErrBadRequest
		}
		t.Completed = strconv.FormatBool(completed)
	}

	for _, date := range []string{t.DueDateMin, t.DueDateMax} {
		if date != "" {
			if _, err := time.Parse("2006-01-02", date); err != nil {
				return ErrBadRequest
			}
		}
	}

	if t.DueDateMin != "" && t.DueDateMax != "" && t.DueDateMin > t.DueDateMax {
		return ErrBadRequest
	}

	if t.OrderBy != "" {
		switch t.OrderBy {
		case "dueDate", "priority", "title", "completed":
		default:
			return ErrBadRequest
		}
	}

	return nil
}

func (t *TaskDTO) ToTask() (Task, error) {
	if strings.TrimSpace(t.DueDate) == "" {
		return Task{
			Title:       t.Title,
			Description: t.Description,
			Priority:    t.Priority,
			Completed:   t.Completed,
		}, nil
	}

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

func ValidateTitle(title string) error {
	if len(strings.TrimSpace(title)) < 2 || len(strings.TrimSpace(title)) > 50 {
		return ErrBadRequest
	}

	return nil
}
