package tests

import (
	"testing"
	"time"

	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/domain"
	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/validation"
	"github.com/stretchr/testify/assert"
)

func TestSanitizeAndValidateTaskTrimsFieldsAndDefaultsPriority(t *testing.T) {
	result := validation.SanitizeAndValidateTask(domain.Task{
		Title:       "  Task title  ",
		Description: "  description  ",
	}, time.Now())

	assert.True(t, result.Valid)
	assert.Equal(t, "Task title", result.Sanitized.Title)
	assert.Equal(t, "description", result.Sanitized.Description)
	assert.Equal(t, 1, result.Sanitized.Priority)
}

func TestSanitizeAndValidateTaskRejectsEmptyTitle(t *testing.T) {
	result := validation.SanitizeAndValidateTask(domain.Task{Title: "   "}, time.Now())

	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors, validation.ErrTaskTitleEmpty)
}

func TestSanitizeAndValidateTaskRejectsInvalidPriority(t *testing.T) {
	result := validation.SanitizeAndValidateTask(domain.Task{Title: "Task", Priority: 4}, time.Now())

	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors, validation.ErrTaskPriorityRange)
}

func TestValidateCreateTaskRequiresUserAndRejectsID(t *testing.T) {
	withID := validation.ValidateCreateTask(domain.Task{ID: 1, UserID: 7, Title: "Task"})
	withoutUser := validation.ValidateCreateTask(domain.Task{Title: "Task"})

	assert.False(t, withID.Valid)
	assert.False(t, withoutUser.Valid)
}

func TestValidateUpdateTaskRequiresID(t *testing.T) {
	result := validation.ValidateUpdateTask(domain.Task{UserID: 7, Title: "Task"})

	assert.False(t, result.Valid)
}
