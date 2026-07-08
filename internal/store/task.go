package store

import (
	"backendGo/internal/entities"
	"database/sql"
	"fmt"
	"time"
)

type SQLiteTaskRepository struct {
	DB *sql.DB
}

func NewSQLiteTaskRepository(db *sql.DB) *SQLiteTaskRepository {
	return &SQLiteTaskRepository{DB: db}
}

func (r *SQLiteTaskRepository) CreateTask(task entities.Task) (entities.Task, error) {
	result, err := r.DB.Exec("INSERT INTO tasks (Tittle, Description, Priority, Completed, DueDate, Tags) VALUES (?,?,?,?,?,?)",
		task.Title, task.Description, task.Priority, task.Completed, task.DueDate, task.Tags)
	if err != nil {
		return entities.Task{}, err
	}

	task.ID, _ = result.LastInsertId()
	return task, nil
}

func (r *SQLiteTaskRepository) GetTasks(searchTerm string, limit int) ([]entities.Task, error) {
	query := "SELECT id, date, title, comment, repeat FROM scheduler"
	args := []interface{}{}

	parsedDate, dateErr := time.Parse("02.01.2006", searchTerm)
	switch {
	case dateErr == nil:
		formattedDate := parsedDate.Format("20060102")
		query += " WHERE date = ? ORDER BY date LIMIT ?"
		args = append(args, formattedDate, limit)
	case searchTerm != "":
		query += " WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?"
		searchTerm = "%" + searchTerm + "%"
		args = append(args, searchTerm, searchTerm, limit)
	default:
		query += " ORDER BY date LIMIT ?"
		args = append(args, limit)
	}

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []entities.Task
	for rows.Next() {
		var task entities.Task
		err = rows.Scan(&task.ID, &task.DueDate, &task.Title, &task.Description)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *SQLiteTaskRepository) GetTaskByID(id int) (entities.Task, error) {
	result, err := r.DB.Query("SELECT * FROM tasks WHERE ID = ?", id)

	if err != nil {
		return entities.Task{}, err
	}

	defer result.Close()

	if result.Next() {
		var task entities.Task
		err := result.Scan(&task.ID, &task.Title, &task.Description, &task.Priority, &task.DueDate, &task.Tags)
		if err != nil {
			return entities.Task{}, err
		}
		return task, nil
	}

	return entities.Task{}, nil
}

func (r *SQLiteTaskRepository) UpdateTask(taskUpdates map[string]entities.Task) (int64, error) {
	query := "UPDATE users SET "
	args := []interface{}{}
	i := 0

	//Loop through all of the changes and add them to the querry
	for key, value := range taskUpdates {
		if key != "id" {
			if i > 0 {
				query += ", "
			}
			query += fmt.Sprintf("%s = ?", key)
			args = append(args, value)
			i++
		}
	}

	//Add the Where statement
	query += " WHERE id = ?"
	args = append(args, taskUpdates["ID"])

	//Execute
	result, err := r.DB.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (r *SQLiteTaskRepository) DeleteTask(id string) (int64, error) {
	result, err := r.DB.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (r *SQLiteTaskRepository) MarkTaskAsDone(id, date string) error {
	_, err := r.DB.Exec("UPDATE scheduler SET date = ? WHERE id = ?", date, id)
	return err
}
