package store

import (
	"backendGo/internal/entities"
	"database/sql"
)

type SQLiteTaskRepository struct {
	DB *sql.DB
}

func NewSQLiteTaskRepository(db *sql.DB) *SQLiteTaskRepository {
	return &SQLiteTaskRepository{DB: db}
}

func (r *SQLiteTaskRepository) CreateTask(task entities.Task) (entities.Task, error) {
	// Here you would implement the logic to insert the task into the SQLite database.
	// For demonstration purposes, we'll just return the task as is.
	return task, nil
}

func (r *SQLiteTaskRepository) GetTaskByID(id int) (entities.Task, error) {
	// Here you would implement the logic to retrieve a task by its ID from the SQLite database.
	// For demonstration purposes, we'll just return a default task and no error.
	return entities.Task{}, nil
}

func (r *SQLiteTaskRepository) UpdateTask(task entities.Task) (entities.Task, error) {
	// Here you would implement the logic to update a task in the SQLite database.
	// For demonstration purposes, we'll just return the task as is.
	return task, nil
}

func (r *SQLiteTaskRepository) DeleteTask(id int) error {
	// Here you would implement the logic to delete a task by its ID from the SQLite database.
	// For demonstration purposes, we'll just return no error.
	return nil
}
