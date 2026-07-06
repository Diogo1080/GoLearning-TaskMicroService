package entities

type Task struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Priority    int      `json:"priority"`
	Completed   bool     `json:"completed"`
	DueDate     string   `json:"due_date,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}
