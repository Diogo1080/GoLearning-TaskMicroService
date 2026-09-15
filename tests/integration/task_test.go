package tests

import (
	"fmt"
	"net/http"
	"os"
	"testing"
)

// ============================================================
// TEST SETUP
// ============================================================

func TestMain(m *testing.M) {
	if err := waitForTaskApp(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// ============================================================
// CREATE TASK
// ============================================================

func TestIntegration_CreateTask_ReturnsCreatedDTO(t *testing.T) {
	_, auth := createTestUser(t)

	status, body := doTaskRequest(
		t,
		http.MethodPost,
		"/task",
		auth.AccessToken,
		taskInsert{
			Title:       "write tests",
			Description: "integration layer",
			Priority:    3,
			DueDate:     "2026-07-09T01:00:00+01:00",
		},
	)

	if status != http.StatusCreated {
		t.Fatalf(
			"Expected 201, got %d: %s",
			status,
			string(body),
		)
	}

	task := decodeTask(
		t,
		body,
	)

	if task.ID == 0 {
		t.Fatal("Expected non-zero task ID")
	}

	if task.Title != "write tests" {
		t.Errorf(
			"Expected title %q, got %q",
			"write tests",
			task.Title,
		)
	}

	if task.Priority != 3 {
		t.Errorf(
			"Expected priority %b, got %b",
			3,
			task.Priority,
		)
	}

	if task.Completed {
		t.Error("New task should not be completed")
	}
}

func TestIntegration_CreateTask_AssignsAuthenticatedUser(t *testing.T) {
	_, auth := createTestUser(t)

	seedTaskViaAPI(
		t,
		auth,
		TaskResponse{Title: "mine only"},
	)

	// Another user must not see it.
	_, otherAuth := createTestUser(t)

	status, body := doTaskRequest(
		t,
		http.MethodGet,
		"/task",
		otherAuth.AccessToken,
		nil,
	)

	if status != http.StatusOK {
		t.Fatalf(
			"Expected 200, got %d: %s",
			status,
			string(body),
		)
	}

	tasks := decodeTaskList(
		t,
		body,
	)

	if len(tasks) != 0 {
		t.Errorf(
			"Expected 0 tasks for unrelated user, got %d",
			len(tasks),
		)
	}
}

func TestIntegration_CreateTask_InvalidBody(t *testing.T) {
	_, auth := createTestUser(t)

	status, body := doTaskRequest(
		t,
		http.MethodPost,
		"/task",
		auth.AccessToken,
		map[string]string{"not": "valid"},
	)

	if status != http.StatusBadRequest {
		t.Fatalf(
			"Expected 400, got %d: %s",
			status,
			string(body),
		)
	}
}

// ============================================================
// GET TASKS
// ============================================================

func TestIntegration_GetTasks_ReturnsOnlyOwnTasks(t *testing.T) {
	_, auth := createTestUser(t)
	seedTaskViaAPI(
		t,
		auth,
		TaskResponse{Title: "task one"},
	)

	seedTaskViaAPI(
		t,
		auth,
		TaskResponse{Title: "task two"},
	)

	// Different user with a different task.
	_, otherAuth := createTestUser(t)
	seedTaskViaAPI(
		t,
		otherAuth,
		TaskResponse{Title: "not yours"},
	)

	status, body := doTaskRequest(
		t,
		http.MethodGet,
		"/task",
		auth.AccessToken,
		nil,
	)

	if status != http.StatusOK {
		t.Fatalf(
			"Expected 200, got %d: %s",
			status,
			string(body),
		)
	}

	tasks := decodeTaskList(
		t,
		body,
	)

	if len(tasks) != 2 {
		t.Fatalf(
			"Expected 2 tasks, got %d",
			len(tasks),
		)
	}

	for _, task := range tasks {
		if task.Title == "not yours" {
			t.Error("User can see another user's task")
		}
	}
}

func TestIntegration_GetTasks_EmptyResultIsArray(t *testing.T) {
	_, auth := createTestUser(t)

	status, body := doTaskRequest(
		t,
		http.MethodGet,
		"/task",
		auth.AccessToken,
		nil,
	)

	if status != http.StatusOK {
		t.Fatalf(
			"Expected 200, got %d: %s",
			status,
			string(body),
		)
	}

	if string(body) == "null" {
		t.Fatal("Expected empty array, got null")
	}

	tasks := decodeTaskList(
		t,
		body,
	)

	if len(tasks) != 0 {
		t.Errorf(
			"Expected 0 tasks, got %d",
			len(tasks),
		)
	}
}

func TestIntegration_GetTasks_SearchFilter(t *testing.T) {
	_, auth := createTestUser(t)

	seedTaskViaAPI(
		t,
		auth,
		TaskResponse{Title: "buy milk"})

	seedTaskViaAPI(
		t,
		auth,
		TaskResponse{Title: "walk dog"})

	status, body := doTaskRequest(
		t,
		http.MethodGet,
		"/task?search=milk",
		auth.AccessToken,
		nil,
	)

	if status != http.StatusOK {
		t.Fatalf(
			"Expected 200, got %d: %s",
			status,
			string(body),
		)
	}

	tasks := decodeTaskList(
		t,
		body,
	)

	if len(tasks) != 1 {
		t.Fatalf(
			"Expected 1 matching task, got %d: %s",
			len(tasks),
			string(body),
		)
	}

	if tasks[0].Title != "buy milk" {
		t.Errorf(
			"Expected title %q, got %q",
			"buy milk",
			tasks[0].Title,
		)
	}
}

// NOTE: this test documents the known query-param typo
// ("dueDataMin" instead of "dueDateMin") in the handler.
// Until that is fixed, date-range filtering silently does
// nothing and this test is expected to FAIL.
func TestIntegration_GetTasks_DateRangeFilter(t *testing.T) {
	_, auth := createTestUser(t)

	seedTaskViaAPI(
		t,
		auth,

		TaskResponse{
			Title:   "in range",
			DueDate: "2026-09-15T01:00:00+01:00",
		},
	)

	seedTaskViaAPI(
		t,
		auth,
		TaskResponse{
			Title:   "out of range",
			DueDate: "2027-01-01T01:00:00+01:00",
		},
	)

	status, body := doTaskRequest(
		t,
		http.MethodGet,
		"/task?dueDateMin=2026-09-01&dueDateMax=2026-09-30",
		auth.AccessToken,
		nil,
	)

	if status != http.StatusOK {
		t.Fatalf(
			"Expected 200, got %d: %s",
			status,
			string(body),
		)
	}

	tasks := decodeTaskList(
		t,
		body,
	)

	if len(tasks) != 1 || tasks[0].Title != "in range" {
		t.Errorf(
			"Expected only 'in range' task, got %s",
			string(body),
		)
	}
}

// ============================================================
// GET TASK BY ID
// ============================================================

func TestIntegration_GetTaskByID(t *testing.T) {
	_, auth := createTestUser(t)

	created := seedTaskViaAPI(
		t,
		auth,
		TaskResponse{
			Title:    "target",
			Priority: 2,
		},
	)

	status, body := doTaskRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/task/%d", created.ID),
		auth.AccessToken,
		nil,
	)

	if status != http.StatusOK {
		t.Fatalf(
			"Expected 200, got %d: %s",
			status,
			string(body),
		)
	}

	task := decodeTask(
		t,
		body,
	)

	if task.ID != created.ID {
		t.Errorf(
			"Expected ID %d, got %d",
			created.ID,
			task.ID,
		)
	}

	if task.Title != "target" {
		t.Errorf(
			"Expected title %q, got %q",
			"target",
			task.Title,
		)
	}
}

func TestIntegration_GetTaskByID_OtherUsersTaskNotFound(t *testing.T) {
	_, creatorAuth := createTestUser(t)

	created := seedTaskViaAPI(
		t,
		creatorAuth,
		TaskResponse{
			Title: "secret",
		},
	)

	_, attackerAuth := createTestUser(t)

	status, body := doTaskRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/task/%d", created.ID),
		attackerAuth.AccessToken,
		nil,
	)

	if status != http.StatusNotFound {
		t.Errorf(
			"Expected 404 for cross-user access, got %d: %s",
			status,
			string(body),
		)
	}
}

func TestIntegration_GetTaskByID_NotFound(t *testing.T) {
	_, auth := createTestUser(t)

	status, body := doTaskRequest(
		t,
		http.MethodGet,
		"/task/999999999",
		auth.AccessToken,
		nil,
	)

	if status != http.StatusNotFound {
		t.Fatalf(
			"Expected 404, got %d: %s",
			status,
			string(body),
		)
	}
}

func TestIntegration_GetTaskByID_InvalidID(t *testing.T) {
	_, auth := createTestUser(t)

	status, body := doTaskRequest(
		t,
		http.MethodGet,
		"/task/not-a-number",
		auth.AccessToken,
		nil,
	)

	if status != http.StatusBadRequest {
		t.Errorf(
			"Expected 400, got %d: %s",
			status,
			string(body),
		)
	}
}

// ============================================================
// UPDATE TASK
// ============================================================

func TestIntegration_UpdateTask_PersistsChanges(t *testing.T) {
	_, auth := createTestUser(t)

	created := seedTaskViaAPI(
		t,
		auth,

		TaskResponse{
			Title:    "target",
			Priority: 1,
		},
	)

	status, body := doTaskRequest(
		t,
		http.MethodPut,
		fmt.Sprintf("/task/%d", created.ID),
		auth.AccessToken,
		TaskResponse{
			Title:       "new title",
			Description: "updated description",
			Priority:    1,
			DueDate:     "2026-10-01T01:00:00+01:00",
		},
	)

	if status != http.StatusOK {
		t.Fatalf(
			"Expected 200, got %d: %s",
			status,
			string(body),
		)
	}

	// Read back through a fresh request to verify persistence.
	status, body = doTaskRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/task/%d", created.ID),
		auth.AccessToken,
		nil,
	)

	if status != http.StatusOK {
		t.Fatalf(
			"Expected 200 on read-back, got %d: %s",
			status,
			string(body),
		)
	}

	task := decodeTask(
		t,
		body,
	)

	if task.Title != "new title" {
		t.Errorf(
			"Expected title %q, got %q",
			"new title",
			task.Title,
		)
	}

	if task.Priority != 1 {
		t.Errorf(
			"Expected priority %b, got %b",
			1,
			task.Priority,
		)
	}

	if task.Description != "updated description" {
		t.Errorf(
			"Expected description %q, got %q",
			"updated description",
			task.Description,
		)
	}
}

func TestIntegration_UpdateTask_OtherUsersTaskNotFound(t *testing.T) {
	_, creatorAuth := createTestUser(t)

	created := seedTaskViaAPI(
		t,
		creatorAuth,

		TaskResponse{
			Title: "not yours",
		},
	)

	_, attackerAuth := createTestUser(t)

	status, body := doTaskRequest(
		t,
		http.MethodPut,
		fmt.Sprintf("/task/%d", created.ID),
		attackerAuth.AccessToken,
		map[string]string{"title": "hijacked"},
	)

	if status != http.StatusNotFound {
		t.Errorf(
			"Expected 404 for cross-user update, got %d: %s",
			status,
			string(body),
		)
	}
}

func TestIntegration_UpdateTask_NotFound(t *testing.T) {
	_, auth := createTestUser(t)

	status, body := doTaskRequest(
		t,
		http.MethodPut,
		"/task/999999999",
		auth.AccessToken,
		map[string]string{"title": "ghost"},
	)

	if status != http.StatusNotFound {
		t.Fatalf(
			"Expected 404, got %d: %s",
			status,
			string(body),
		)
	}
}

// ============================================================
// COMPLETE TASK
// ============================================================

func TestIntegration_CompleteTask_SetsCompleted(t *testing.T) {
	_, auth := createTestUser(t)

	created := seedTaskViaAPI(
		t,
		auth,
		TaskResponse{
			Title: "finish me",
		},
	)

	status, body := doTaskRequest(
		t,
		http.MethodPatch,
		fmt.Sprintf("/task/complete/%d", created.ID),
		auth.AccessToken,
		nil,
	)

	if status != http.StatusOK {
		t.Fatalf(
			"Expected 200, got %d: %s",
			status,
			string(body),
		)
	}

	// Read back and confirm completed flag persisted.
	status, body = doTaskRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/task/%d", created.ID),
		auth.AccessToken,
		nil,
	)

	task := decodeTask(
		t,
		body,
	)

	if status != http.StatusOK || !task.Completed {
		t.Errorf(
			"Expected task to be completed, got status=%d completed=%v",
			status,
			task.Completed,
		)
	}
}

func TestIntegration_CompleteTask_OtherUsersTaskNotFound(t *testing.T) {
	_, creatorAuth := createTestUser(t)

	created := seedTaskViaAPI(
		t,
		creatorAuth,

		TaskResponse{
			Title: "someone eles's",
		},
	)

	_, attackerAuth := createTestUser(t)

	status, body := doTaskRequest(
		t,
		http.MethodPatch,
		fmt.Sprintf("/task/%d/complete", created.ID),
		attackerAuth.AccessToken,
		nil,
	)

	if status != http.StatusNotFound {
		t.Errorf(
			"Expected 404 for cross-user complete, got %d: %s",
			status,
			string(body),
		)
	}
}

// ============================================================
// DELETE TASK
// ============================================================

func TestIntegration_DeleteTask_RemovesTask(t *testing.T) {
	_, auth := createTestUser(t)

	created := seedTaskViaAPI(
		t,
		auth,

		TaskResponse{
			Title: "delete me",
		},
	)

	status, body := doTaskRequest(
		t,
		http.MethodDelete,
		fmt.Sprintf("/task/%d", created.ID),
		auth.AccessToken,
		nil,
	)

	if status != http.StatusOK {
		t.Fatalf(
			"Expected 200, got %d: %s",
			status,
			string(body),
		)
	}

	// The record must be gone from the database.
	status, body = doTaskRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/task/%d", created.ID),
		auth.AccessToken,
		nil,
	)

	if status != http.StatusNotFound {
		t.Fatalf(
			"Expected 404 after deletion, got %d: %s",
			status,
			string(body),
		)
	}
}

func TestIntegration_DeleteTask_TwiceSecondIsNotFound(t *testing.T) {
	_, auth := createTestUser(t)

	created := seedTaskViaAPI(
		t,
		auth,

		TaskResponse{
			Title: "once",
		},
	)

	doTaskRequest(
		t,
		http.MethodDelete,
		fmt.Sprintf("/task/%d", created.ID),
		auth.AccessToken,
		nil,
	)

	status, body := doTaskRequest(
		t,
		http.MethodDelete,
		fmt.Sprintf("/task/%d", created.ID),
		auth.AccessToken,
		nil,
	)

	if status != http.StatusNotFound {
		t.Errorf(
			"Expected 404 on second delete, got %d: %s",
			status,
			string(body),
		)
	}
}

func TestIntegration_DeleteTask_OtherUsersTaskNotFound(t *testing.T) {
	_, creatorAuth := createTestUser(t)

	created := seedTaskViaAPI(
		t,
		creatorAuth,

		TaskResponse{
			Title: "protected",
		},
	)

	_, attackerAuth := createTestUser(t)

	status, body := doTaskRequest(
		t,
		http.MethodDelete,
		fmt.Sprintf("/task/%d", created.ID),
		attackerAuth.AccessToken,
		nil,
	)

	if status != http.StatusNotFound {
		t.Errorf(
			"Expected 404 for cross-user delete, got %d: %s",
			status,
			string(body),
		)
	}

	// Original owner must still have the task.
	status, _ = doTaskRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/task/%d", created.ID),
		creatorAuth.AccessToken,
		nil,
	)

	if status != http.StatusOK {
		t.Error("Owner lost access to task after someone else's failed delete attempt")
	}
}

// ============================================================
// PERSISTENCE ACROSS REQUESTS
// ============================================================

func TestIntegration_TaskPersistsAcrossRequests(t *testing.T) {
	_, auth := createTestUser(t)

	created := seedTaskViaAPI(
		t,
		auth,
		TaskResponse{
			Title:    "persistent",
			Priority: 1,
		},
	)

	// Two independent GET requests must return identical data.
	first := getTaskAssertOK(
		t,
		auth,
		created.ID,
	)

	second := getTaskAssertOK(
		t,
		auth,
		created.ID,
	)

	if first.Title != second.Title {
		t.Fatalf(
			"Title changed between requests: %q != %q",
			first.Title,
			second.Title,
		)
	}

	if first.Priority != second.Priority {
		t.Fatalf(
			"Priority changed between requests: %b != %b",
			first.Priority,
			second.Priority,
		)
	}

	if first.ID != second.ID {
		t.Fatalf(
			"ID changed between requests: %d != %d",
			first.ID,
			second.ID,
		)
	}
}

// ============================================================
// FULL LIFECYCLE
// ============================================================

func TestIntegration_TaskFullLifecycle(t *testing.T) {
	_, auth := createTestUser(t)

	// Create.
	created := seedTaskViaAPI(
		t,
		auth,

		TaskResponse{
			Title:    "lifecycle",
			Priority: 2,
		},
	)

	// Update.
	status, body := doTaskRequest(
		t,
		http.MethodPut,
		fmt.Sprintf("/task/%d", created.ID),
		auth.AccessToken,
		TaskResponse{
			Title:    "lifecycle updated",
			Priority: 3,
		},
	)

	if status != http.StatusOK {
		t.Fatalf(
			"Update failed: status=%d body=%s",
			status,
			string(body),
		)
	}

	// Complete.
	status, body = doTaskRequest(
		t,
		http.MethodPatch,
		fmt.Sprintf("/task/complete/%d", created.ID),
		auth.AccessToken,
		nil,
	)

	if status != http.StatusOK {
		t.Fatalf(
			"Complete failed: status=%d body=%s",
			status,
			string(body),
		)
	}

	// Verify final state.
	final := getTaskAssertOK(
		t,
		auth,
		created.ID,
	)

	if final.Title != "lifecycle updated" || final.Priority != 3 || !final.Completed {
		t.Errorf(
			"Unexpected final state: %+v",
			final,
		)
	}

	// Delete.
	status, _ = doTaskRequest(
		t,
		http.MethodDelete,
		fmt.Sprintf("/task/%d", created.ID),
		auth.AccessToken,
		nil,
	)

	if status != http.StatusOK {
		t.Errorf(
			"Delete failed: status=%d",
			status,
		)
	}
}
