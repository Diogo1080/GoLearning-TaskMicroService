package domain

import (
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
	Offset     int
}

type Task struct {
	ID          int64
	UserID      int64
	Title       string
	Description string
	Priority    int
	Completed   bool
	DueDate     time.Time
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
	if t.Offset <= 0 {
		t.Offset = 1
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

func ValidateTitle(title string) error {
	if len(strings.TrimSpace(title)) < 2 || len(strings.TrimSpace(title)) > 50 {
		return ErrBadRequest
	}

	return nil
}
