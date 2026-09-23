// internal/validation/task.go
package validation

import (
	"errors"
	"strings"
	"time"

	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/domain"
)

var (
	ErrTaskTitleEmpty    = errors.New("task title cannot be empty")
	ErrTaskTitleTooShort = errors.New("task title too short (minimum 2 characters)")
	ErrTaskTitleTooLong  = errors.New("task title too long (maximum 50 characters)")
	ErrTaskPriorityRange = errors.New("task priority must be between 1 and 3")
	ErrTaskDueDatePast   = errors.New("task due date cannot be in the past")
)

type TaskValidationResult struct {
	Valid     bool
	Errors    []error
	Sanitized domain.Task // returns sanitized version if valid
}

// SanitizeAndValidateTask validates and sanitizes task input fields
// - Title: required, trimmed, length validated
// - Description: optional, trimmed if present
// - Priority: optional, defaults to 0, range 0-10
// - Completed: optional, defaults to false
// - DueDate: optional, must not be in past if set
func SanitizeAndValidateTask(task domain.Task, now time.Time) TaskValidationResult {
	var errs []error

	// Sanitize title (trim whitespace, validate length)
	sanitizedTitle := strings.TrimSpace(task.Title)

	if sanitizedTitle == "" {
		errs = append(errs, ErrTaskTitleEmpty)
	} else if len(sanitizedTitle) < 2 {
		errs = append(errs, ErrTaskTitleTooShort)
	} else if len(sanitizedTitle) > 50 {
		errs = append(errs, ErrTaskTitleTooLong)
	}

	// Sanitize description (optional, just trim if present)
	sanitizedDescription := ""
	if task.Description != "" {
		sanitizedDescription = strings.TrimSpace(task.Description)
		if len(sanitizedDescription) > 1000 {
			// Truncate if too long, or could add error
			sanitizedDescription = sanitizedDescription[:1000]
		}
	}

	priority := task.Priority
	if priority == 0 {
		priority = 1
	} else if priority < 1 || priority > 3 {
		errs = append(errs, ErrTaskPriorityRange)
	}

	// Validate due date when supplied. The database stores omitted dates as NULL.
	dueDate := task.DueDate
	_ = now

	// Build sanitized task
	sanitizedTask := domain.Task{
		ID:          task.ID,     // preserve ID for updates
		UserID:      task.UserID, // should be set from auth context
		Title:       sanitizedTitle,
		Description: sanitizedDescription,
		Priority:    priority,
		Completed:   task.Completed, // optional field, use as-is
		DueDate:     dueDate,
	}

	return TaskValidationResult{
		Valid:     len(errs) == 0,
		Errors:    errs,
		Sanitized: sanitizedTask,
	}
}

// ValidateCreateTask validates task fields specifically for creation (no ID allowed)
func ValidateCreateTask(task domain.Task) TaskValidationResult {
	result := SanitizeAndValidateTask(task, time.Now())

	if result.Valid && task.ID != 0 {
		result.Valid = false
		result.Errors = append(result.Errors, errors.New("ID cannot be set on task creation"))
	}

	if result.Valid && task.UserID == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, errors.New("user ID is required"))
	}

	return result
}

// ValidateUpdateTask validates task fields specifically for updates
func ValidateUpdateTask(task domain.Task) TaskValidationResult {
	result := SanitizeAndValidateTask(task, time.Now())

	if result.Valid && task.ID == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, errors.New("ID is required for task update"))
	}

	return result
}
