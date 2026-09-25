package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

// TestUser holds credentials for test users (stored locally, not from API)
type TestUser struct {
	Username  string
	Email     string
	Password  string
	Birthdate string
	UserID    float64 // populated after registration
}

type taskInsert struct {
	Title       string
	Description string
	Priority    int
	DueDate     string
}

// RegisterResponse matches the actual register API response
type RegisterResponse struct {
	Message string  `json:"message"`
	UserID  float64 `json:"user_id"`
}

// AuthResponse matches the actual login API response
type AuthResponse struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	UserID       float64 `json:"user_id"`
	Message      string  `json:"message"`
}

type UserResponse struct {
	ID        int32  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Birthdate string `json:"birthdate"`
}

// Task service (under test) and Identity service (provides users/tokens).
var taskBaseURL = "http://localhost:9003/api"
var identityBaseURL = "http://localhost:9001/api"

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ============================================================
// TYPES
// ============================================================

type TaskResponse struct {
	ID          int    `json:"id"`
	UserID      int    `json:"userId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
	Completed   bool   `json:"completed"`
	DueDate     string `json:"dueDate,omitempty"`
}

// waitForTaskApp waits for the task service to be ready.
func waitForTaskApp() error {
	for i := 0; i < 30; i++ {
		resp, err := http.Get(taskBaseURL + "/health")

		if err == nil {
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		time.Sleep(time.Second)
	}

	return fmt.Errorf(
		"task service did not become ready at %s",
		taskBaseURL,
	)
}

// doTaskRequest makes an authenticated request against the task service.
func doTaskRequest(t *testing.T, method, path, accessToken string, body interface{}) (int, []byte) {
	t.Helper()

	var reqBody io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, taskBaseURL+path, reqBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, respBody
}

// createTask registers and logs in a user, then creates a task.
// Returns the auth tokens and the created task.
func createTaskForTestUser(t *testing.T, title string, priority string) (AuthResponse, TaskResponse) {
	t.Helper()

	_, auth := createTestUser(t)

	status, body := doTaskRequest(
		t,
		http.MethodPost,
		"/tasks",
		auth.AccessToken,
		map[string]string{
			"title":    title,
			"priority": priority,
		},
	)

	if status != http.StatusCreated {
		t.Fatalf("Create task failed: status=%d body=%s", status, string(body))
	}

	var task TaskResponse
	json.Unmarshal(body, &task)

	return auth, task
}

// seedTaskViaAPI creates a task with explicit fields via the API.
func seedTaskViaAPI(t *testing.T, auth AuthResponse, input TaskResponse) TaskResponse {
	t.Helper()

	status, body := doTaskRequest(
		t,
		http.MethodPost,
		"/task",
		auth.AccessToken,
		input,
	)

	if status != http.StatusCreated {
		t.Fatalf(
			"Seed task failed: status=%d body=%s",
			status,
			string(body),
		)
	}

	var task TaskResponse
	json.Unmarshal(body, &task)
	return task
}

func createTestUser(t *testing.T) (TestUser, AuthResponse) {
	t.Helper()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	user := TestUser{
		Username:  "db_user_" + suffix,
		Email:     "db" + suffix + "@test.com",
		Password:  "SecurePass123!",
		Birthdate: "1990-05-15",
	}

	registeredUser := registerUser(t, user)

	auth := loginUser(
		t,
		registeredUser.Username,
		registeredUser.Password,
	)

	if auth.UserID == 0 {
		t.Fatal("Expected non-zero user ID")
	}

	user.UserID = auth.UserID

	return user, auth
}

// registerUser registers a new user and returns the response
func registerUser(t *testing.T, user TestUser) TestUser {
	t.Helper()

	input := map[string]string{
		"username":  user.Username,
		"password":  user.Password,
		"email":     user.Email,
		"birthdate": user.Birthdate,
	}

	body, _ := json.Marshal(input)
	resp, err := http.Post(identityBaseURL+"/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Register failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var regResp RegisterResponse
	json.Unmarshal(respBody, &regResp)

	// Return as AuthResponse format for consistency
	return TestUser{
		Username:  user.Username,
		Email:     user.Email,
		Password:  user.Password,
		Birthdate: user.Birthdate,
		UserID:    regResp.UserID,
	}
}

// loginUser logs in and returns the auth tokens
func loginUser(t *testing.T, identifier, password string) AuthResponse {
	t.Helper()

	input := map[string]string{
		"usernameoremail": identifier, // accepts username OR email
		"password":        password,
	}

	body, _ := json.Marshal(input)
	resp, err := http.Post(identityBaseURL+"/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Login failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var result AuthResponse
	json.Unmarshal(respBody, &result)
	return result
}

// decodeTask decodes a single task response body.
func decodeTask(t *testing.T, body []byte) TaskResponse {
	t.Helper()

	var task TaskResponse
	if err := json.Unmarshal(body, &task); err != nil {
		t.Fatalf("Failed to decode task response: %v", err)
	}
	return task
}

// decodeTaskList decodes a task list response body.
func decodeTaskList(t *testing.T, body []byte) []TaskResponse {
	t.Helper()

	var tasks []TaskResponse
	if err := json.Unmarshal(body, &tasks); err != nil {
		t.Fatalf("Failed to decode task list response: %v", err)
	}
	return tasks
}

// getTaskAssertOK fetches a task and fails the test on non-200.
func getTaskAssertOK(t *testing.T, auth AuthResponse, id int) TaskResponse {
	t.Helper()

	status, body := doTaskRequest(
		t,
		http.MethodGet,
		fmt.Sprintf("/task/%d", id),
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

	return decodeTask(
		t,
		body,
	)
}
