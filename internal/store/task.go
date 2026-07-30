package store

import (
	entities "backendGo/internal/domain"
	"database/sql"
	"fmt"
)

type SQLiteTaskRepository struct {
	DB *sql.DB
}

func NewSQLiteTaskRepository(db *sql.DB) *SQLiteTaskRepository {
	return &SQLiteTaskRepository{DB: db}
}

func (r *SQLiteTaskRepository) CreateTask(task entities.Task) (entities.Task, error) {
	result, err := r.DB.Exec("INSERT INTO tasks (user_id, title, description, priority, completed, dueDate) VALUES (?,?,?,?,?,?)",
		task.UserID, task.Title, task.Description, task.Priority, task.Completed, task.DueDate)

	if err != nil {
		return entities.Task{}, err
	}

	task.ID, _ = result.LastInsertId()
	return task, nil
}

func (r *SQLiteTaskRepository) GetTasks(terms entities.TaskSearch) ([]entities.Task, error) {
	query := "SELECT * FROM tasks "
	args := []interface{}{}

	query += "WHERE user_id = ? "
	args = append(args, terms.UserID)

	if terms.Search != "" {
		query += "AND (title LIKE ? OR description LIKE ?) "
		args = append(args, "%"+terms.Search+"%", "%"+terms.Search+"%")
	}

	if terms.Priority != "" {
		query += "AND priority = ? "
		args = append(args, terms.Priority)
	}

	if terms.Completed != "" {
		query += "AND completed = ? "
		args = append(args, terms.Completed)
	}

	if terms.DueDateMax != "" {
		query += "AND dueDate < ? "
		args = append(args, terms.DueDateMax)
	}

	if terms.DueDateMin != "" {
		query += "AND dueDate > ? "
		args = append(args, terms.DueDateMin)
	}

	if terms.OrderBy != "" {
		query += "ORDER BY ? "
		args = append(args, terms.OrderBy)
	} else {
		query += "ORDER BY dueDate "
	}

	query += "Limit ? "
	args = append(args, terms.Limit)

	fmt.Print(query)
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []entities.Task
	for rows.Next() {
		var task entities.Task
		err = rows.Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Priority, &task.Completed, &task.DueDate)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *SQLiteTaskRepository) GetTaskByID(id int, userID int) (entities.Task, error) {
	result, err := r.DB.Query("SELECT * FROM tasks WHERE id = ? AND  user_id = ?", id, userID)

	if err != nil {
		return entities.Task{}, entities.ErrDatabaseFailed
	}

	defer result.Close()

	if result.Next() {
		var task entities.Task
		err := result.Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Priority, &task.Completed, &task.DueDate)
		if err != nil {
			return entities.Task{}, err
		}
		return task, nil
	}

	return entities.Task{}, entities.ErrNotFound
}

func (r *SQLiteTaskRepository) UpdateTask(taskUpdates entities.Task, taskID int, userID int) (entities.Task, error) {
	result, err := r.DB.Exec("UPDATE tasks SET title = ?, description = ?, Priority = ?, DueDate = ? WHERE id = ? AND user_id = ?",
		taskUpdates.Title, taskUpdates.Description, taskUpdates.Priority, taskUpdates.DueDate, taskID, userID)

	if err != nil {
		return entities.Task{}, err
	}

	affected, err := result.RowsAffected()

	fmt.Print(affected)

	if affected == int64(0) {
		return entities.Task{}, entities.ErrNotFound
	}

	return taskUpdates, nil
}

func (r *SQLiteTaskRepository) DeleteTask(id int, userID int) (int64, error) {
	result, err := r.DB.Exec("DELETE FROM tasks WHERE id = ? AND  user_id = ?", id, userID)

	if err != nil {
		return 0, err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return 0, entities.ErrNotFound
	}

	return result.RowsAffected()
}

func (r *SQLiteTaskRepository) MarkTaskAsDone(id int, userID int) (int64, error) {
	result, err := r.DB.Exec("UPDATE tasks SET completed = ? WHERE id = ? AND  user_id = ?", true, id, userID)
	if err != nil {
		return 0, err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return 0, entities.ErrNotFound
	}

	return result.RowsAffected()
}
