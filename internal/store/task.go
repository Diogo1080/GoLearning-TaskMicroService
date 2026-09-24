package store

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	entities "github.com/Diogo1080/GoLearning-TaskMicroService/internal/domain"
)

type SQLiteTaskRepository struct {
	DB *sql.DB
}

func NewSQLiteTaskRepository(db *sql.DB) *SQLiteTaskRepository {
	return &SQLiteTaskRepository{DB: db}
}

func (r *SQLiteTaskRepository) CreateTask(ctx context.Context, task entities.Task) (entities.Task, error) {
	var dueDate interface{}
	if !task.DueDate.IsZero() {
		dueDate = task.DueDate
	}
	err := r.DB.QueryRowContext(ctx, "INSERT INTO tasks (user_id, title, description, priority, completed, due_date) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		task.UserID, task.Title, task.Description, task.Priority, task.Completed, dueDate).Scan(&task.ID)

	if err != nil {
		return entities.Task{}, err
	}

	return task, nil
}

func (r *SQLiteTaskRepository) GetTasks(ctx context.Context, terms entities.TaskSearch) ([]entities.Task, error) {
	if err := terms.Sanatise(); err != nil {
		return nil, err
	}

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
		priority, _ := strconv.Atoi(terms.Priority)
		args = append(args, priority)
		i++
	}

	if terms.Completed != "" {
		query += "AND completed = $" + fmt.Sprint(i) + " "
		completed, _ := strconv.ParseBool(terms.Completed)
		args = append(args, completed)
		i++
	}

	if terms.DueDateMax != "" {
		query += "AND due_date < $" + fmt.Sprint(i) + " "
		date, _ := time.Parse("2006-01-02", terms.DueDateMax)
		args = append(args, date)
		i++
	}

	if terms.DueDateMin != "" {
		query += "AND due_date > $" + fmt.Sprint(i) + " "
		date, _ := time.Parse("2006-01-02", terms.DueDateMin)
		args = append(args, date)
		i++
	}

	if terms.OrderBy != "" {
		orderBy := map[string]string{
			"dueDate":   "due_date",
			"priority":  "priority",
			"title":     "title",
			"completed": "completed",
		}[terms.OrderBy]
		query += "ORDER BY " + orderBy + " "
	} else {
		query += "ORDER BY due_date "
	}

	query += "Limit $" + fmt.Sprint(i) + " "
	args = append(args, terms.Limit)
	i++

	query += "OFFSET $" + fmt.Sprint(i) + " "
	args = append(args, (terms.Offset-1)*terms.Limit)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []entities.Task
	for rows.Next() {
		var task entities.Task
		var dueDate sql.NullTime
		err = rows.Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Priority, &task.Completed, &dueDate)
		if err != nil {
			return nil, err
		}
		if dueDate.Valid {
			task.DueDate = dueDate.Time
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *SQLiteTaskRepository) GetTaskByID(ctx context.Context, id int, userID int) (entities.Task, error) {
	result, err := r.DB.QueryContext(ctx, "SELECT * FROM tasks WHERE id = $1 AND user_id = $2", id, userID)

	if err != nil {
		return entities.Task{}, entities.ErrInternal
	}

	defer result.Close()

	if result.Next() {
		var task entities.Task
		var dueDate sql.NullTime
		err := result.Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Priority, &task.Completed, &dueDate)
		if err != nil {
			return entities.Task{}, err
		}
		if dueDate.Valid {
			task.DueDate = dueDate.Time
		}
		return task, nil
	}

	return entities.Task{}, entities.ErrNotFound
}

func (r *SQLiteTaskRepository) UpdateTask(ctx context.Context, taskUpdates entities.Task, taskID int, userID int) (entities.Task, error) {
	var dueDate interface{}
	if !taskUpdates.DueDate.IsZero() {
		dueDate = taskUpdates.DueDate
	}
	result, err := r.DB.ExecContext(ctx, "UPDATE tasks SET title = $1, description = $2, priority = $3, due_date = $4 WHERE id = $5 AND user_id = $6",
		taskUpdates.Title, taskUpdates.Description, taskUpdates.Priority, dueDate, taskID, userID)

	if err != nil {
		return entities.Task{}, err
	}

	affected, err := result.RowsAffected()

	if affected == int64(0) {
		return entities.Task{}, entities.ErrNotFound
	}

	return taskUpdates, nil
}

func (r *SQLiteTaskRepository) DeleteTask(ctx context.Context, id int, userID int) (int64, error) {
	result, err := r.DB.ExecContext(ctx, "DELETE FROM tasks WHERE id = $1 AND user_id = $2", id, userID)

	if err != nil {
		return 0, err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return 0, entities.ErrNotFound
	}

	return result.RowsAffected()
}

func (r *SQLiteTaskRepository) MarkTaskAsDone(ctx context.Context, id int, userID int) (int64, error) {
	result, err := r.DB.ExecContext(ctx, "UPDATE tasks SET completed = $1 WHERE id = $2 AND  user_id = $3", true, id, userID)
	if err != nil {
		return 0, err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return 0, entities.ErrNotFound
	}

	return result.RowsAffected()
}
