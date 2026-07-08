package entities

import "strings"

type Task struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Priority    int      `json:"priority"`
	Completed   bool     `json:"completed"`
	DueDate     string   `json:"due_date,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type TaskDTO struct {
	Title       string `form:"title" json:"title"`
	Description string `form:"description" json:"description"`
	Priority    string `form:"priority" json:"priority"`
	Completed   string `form:"completed" json:"completed"`
	DueDate     string `form:"due_date" json:"due_date"`
	Tags        string `form:"tags" json:"tags"`
}

func (t *TaskDTO) ToTask() Task {
	return Task{
		Title:       t.Title,
		Description: t.Description,
		Priority:    parsePriority(t.Priority),
		Completed:   parseCompleted(t.Completed),
		DueDate:     t.DueDate,
		Tags:        parseTags(t.Tags),
	}
}

func parsePriority(priority string) int {
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

func parseTags(tags string) []string {
	if tags == "" {
		return []string{}
	}

	return strings.Split(tags, ",")
}
