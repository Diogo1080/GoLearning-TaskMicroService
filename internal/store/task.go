package store

import (
	"backendGo/internal/entities"
	"database/sql"
	"time"
)

type SQLiteTaskRepository struct {
	DB *sql.DB
}

func NewSQLiteTaskRepository(db *sql.DB) *SQLiteTaskRepository {
	return &SQLiteTaskRepository{DB: db}
}

func (r *SQLiteTaskRepository) CreateTask(task entities.Task) (entities.TaskDTO, error) {
	result, err := r.DB.Exec("INSERT INTO tasks (title, description, priority, completed, dueDate) VALUES (?,?,?,?,?)",
		task.Title, task.Description, task.Priority, task.Completed, task.DueDate)

	if err != nil {
		return entities.TaskDTO{}, err
	}

	task.ID, _ = result.LastInsertId()
	return task.ToTaskDTO()
}

func (r *SQLiteTaskRepository) GetTasks(terms entities.TaskSearch, limit int) ([]entities.Task, error) {
	query := "SELECT * FROM tasks"
	args := []interface{}{}

	parsedDate, dateErr := time.Parse("02.01.2006", terms.Search)
	switch {
	case dateErr == nil:
		formattedDate := parsedDate.Format("20060102")
		query += " WHERE DueDate = ? ORDER BY date LIMIT ?"
		args = append(args, formattedDate, limit)
	case terms.Search != "":
		query += " WHERE title LIKE ? OR description LIKE ? ORDER BY DueDate LIMIT ?"
		terms.Search = "%" + terms.Search + "%"
		args = append(args, terms.Search, terms.Search, limit)
	default:
		query += " ORDER BY DueDate LIMIT ?"
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
		err = rows.Scan(&task.ID, &task.Title, &task.Description, &task.Priority, &task.Completed, &task.DueDate)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *SQLiteTaskRepository) GetTaskByID(id int) (entities.TaskDTO, error) {
	result, err := r.DB.Query("SELECT * FROM tasks WHERE ID = ?", id)

	if err != nil {
		return entities.TaskDTO{}, err
	}

	defer result.Close()

	if result.Next() {
		var task entities.Task
		err := result.Scan(&task.ID, &task.Title, &task.Description, &task.Priority, &task.Completed, &task.DueDate)
		if err != nil {
			return entities.TaskDTO{}, err
		}
		return task.ToTaskDTO()
	}

	return entities.TaskDTO{}, nil
}

func (r *SQLiteTaskRepository) UpdateTask(taskUpdates entities.Task, id int) (entities.TaskDTO, error) {
	result, err := r.DB.Exec("UPDATE tasks SET title = ?, description = ?, Priority = ?, DueDate = ? WHERE id = ?",
		taskUpdates.Title, taskUpdates.Description, taskUpdates.Priority, taskUpdates.DueDate, id)

	if err != nil {
		return entities.TaskDTO{}, err
	}

	_, err = result.RowsAffected()

	return r.GetTaskByID(id)
}

func (r *SQLiteTaskRepository) DeleteTask(id int) (int64, error) {
	result, err := r.DB.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (r *SQLiteTaskRepository) MarkTaskAsDone(id int) (int64, error) {
	result, err := r.DB.Exec("UPDATE tasks SET completed = ? WHERE id = ?", true, id)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
