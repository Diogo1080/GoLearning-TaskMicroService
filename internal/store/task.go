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
	result, err := r.DB.Exec("INSERT INTO tasks (user_id, title, description, priority, completed, dueDate) VALUES ($1, $2, $3, $4, $5, $6)",
		task.UserID, task.Title, task.Description, task.Priority, task.Completed, task.DueDate)

	if err != nil {
		return entities.Task{}, err
	}

	task.ID, _ = result.LastInsertId()
	return task, nil
}

func (r *SQLiteTaskRepository) GetTasks(terms entities.TaskSearch) ([]entities.Task, error) {
	i := int(1)
	query := "SELECT * FROM tasks "
	args := []interface{}{}

	query += "WHERE user_id = $" + fmt.Sprint(i) + " "
	args = append(args, terms.UserID)
	i++

	if terms.Search != "" {
		query += "AND (title LIKE $" + fmt.Sprint(i) + " OR description LIKE $" + fmt.Sprint(i+1) + ") "
		args = append(args, "%"+terms.Search+"%", "%"+terms.Search+"%")
		i += 2
	}

	if terms.Priority != "" {
		query += "AND priority = $" + fmt.Sprint(i) + " "
		args = append(args, terms.Priority)
		i++
	}

	if terms.Completed != "" {
		query += "AND completed = $" + fmt.Sprint(i) + " "
		args = append(args, terms.Completed)
		i++
	}

	if terms.DueDateMax != "" {
		query += "AND dueDate < $" + fmt.Sprint(i) + " "
		args = append(args, terms.DueDateMax)
		i++
	}

	if terms.DueDateMin != "" {
		query += "AND dueDate > $" + fmt.Sprint(i) + " "
		args = append(args, terms.DueDateMin)
		i++
	}

	if terms.OrderBy != "" {
		query += "ORDER BY $" + fmt.Sprint(i) + " "
		args = append(args, terms.OrderBy)
		i++
	} else {
		query += "ORDER BY dueDate "
	}

	query += "Limit $" + fmt.Sprint(i) + " "
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
	result, err := r.DB.Query("SELECT * FROM tasks WHERE id = $1 AND user_id = $2", id, userID)

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
	result, err := r.DB.Exec("UPDATE tasks SET title = $1, description = $2, priority = $3, dueDate = $4 WHERE id = $5 AND user_id = $6",
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
	result, err := r.DB.Exec("DELETE FROM tasks WHERE id = $1 AND user_id = $2", id, userID)

	if err != nil {
		return 0, err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return 0, entities.ErrNotFound
	}

	return result.RowsAffected()
}

func (r *SQLiteTaskRepository) MarkTaskAsDone(id int, userID int) (int64, error) {
	result, err := r.DB.Exec("UPDATE tasks SET completed = $1 WHERE id = $2 AND  user_id = $3", true, id, userID)
	if err != nil {
		return 0, err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return 0, entities.ErrNotFound
	}

	return result.RowsAffected()
}
