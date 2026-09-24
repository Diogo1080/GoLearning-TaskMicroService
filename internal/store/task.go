package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	entities "github.com/Diogo1080/GoLearning-TaskMicroService/internal/domain"
	"github.com/lib/pq"
)

const taskColumns = "id, user_id, title, description, priority, completed, due_date"

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
		return entities.Task{}, mapDatabaseError(err)
	}

	return task, nil
}

func (r *SQLiteTaskRepository) GetTasks(ctx context.Context, terms entities.TaskSearch) ([]entities.Task, error) {
	if err := terms.Sanatise(); err != nil {
		return nil, err
	}

	i := int(1)
	query := "SELECT " + taskColumns + " FROM tasks "
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
		priority, err := strconv.Atoi(terms.Priority)
		if err != nil {
			return nil, entities.ErrBadRequest
		}
		args = append(args, priority)
		i++
	}

	if terms.Completed != "" {
		query += "AND completed = $" + fmt.Sprint(i) + " "
		completed, err := strconv.ParseBool(terms.Completed)
		if err != nil {
			return nil, entities.ErrBadRequest
		}
		args = append(args, completed)
		i++
	}

	if terms.DueDateMax != "" {
		query += "AND due_date < $" + fmt.Sprint(i) + " "
		date, err := time.Parse("2006-01-02", terms.DueDateMax)
		if err != nil {
			return nil, entities.ErrBadRequest
		}
		args = append(args, date)
		i++
	}

	if terms.DueDateMin != "" {
		query += "AND due_date > $" + fmt.Sprint(i) + " "
		date, err := time.Parse("2006-01-02", terms.DueDateMin)
		if err != nil {
			return nil, entities.ErrBadRequest
		}
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
		return nil, mapDatabaseError(err)
	}
	defer rows.Close()

	var tasks []entities.Task
	for rows.Next() {
		var task entities.Task
		var dueDate sql.NullTime
		err = rows.Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Priority, &task.Completed, &dueDate)
		if err != nil {
			return nil, mapDatabaseError(err)
		}
		if dueDate.Valid {
			task.DueDate = dueDate.Time
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, mapDatabaseError(err)
	}

	return tasks, nil
}

func (r *SQLiteTaskRepository) GetTaskByID(ctx context.Context, id int, userID int) (entities.Task, error) {
	var task entities.Task
	var dueDate sql.NullTime
	err := r.DB.QueryRowContext(ctx, "SELECT "+taskColumns+" FROM tasks WHERE id = $1 AND user_id = $2", id, userID).
		Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Priority, &task.Completed, &dueDate)

	if err != nil {
		return entities.Task{}, mapDatabaseError(err)
	}

	if dueDate.Valid {
		task.DueDate = dueDate.Time
	}

	return task, nil
}

func (r *SQLiteTaskRepository) UpdateTask(ctx context.Context, taskUpdates entities.Task, taskID int, userID int) (entities.Task, error) {
	var dueDate interface{}
	if !taskUpdates.DueDate.IsZero() {
		dueDate = taskUpdates.DueDate
	}
	result, err := r.DB.ExecContext(ctx, "UPDATE tasks SET title = $1, description = $2, priority = $3, due_date = $4 WHERE id = $5 AND user_id = $6",
		taskUpdates.Title, taskUpdates.Description, taskUpdates.Priority, dueDate, taskID, userID)

	if err != nil {
		return entities.Task{}, mapDatabaseError(err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return entities.Task{}, mapDatabaseError(err)
	}

	if affected == int64(0) {
		return entities.Task{}, entities.ErrNotFound
	}

	return taskUpdates, nil
}

func (r *SQLiteTaskRepository) DeleteTask(ctx context.Context, id int, userID int) (int64, error) {
	result, err := r.DB.ExecContext(ctx, "DELETE FROM tasks WHERE id = $1 AND user_id = $2", id, userID)

	if err != nil {
		return 0, mapDatabaseError(err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, mapDatabaseError(err)
	}
	if rows == 0 {
		return 0, entities.ErrNotFound
	}

	return rows, nil
}

func (r *SQLiteTaskRepository) MarkTaskAsDone(ctx context.Context, id int, userID int) (int64, error) {
	result, err := r.DB.ExecContext(ctx, "UPDATE tasks SET completed = $1 WHERE id = $2 AND  user_id = $3", true, id, userID)
	if err != nil {
		return 0, mapDatabaseError(err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, mapDatabaseError(err)
	}
	if rows == 0 {
		return 0, entities.ErrNotFound
	}

	return rows, nil
}

func mapDatabaseError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return entities.ErrNotFound
	}

	var postgresError *pq.Error
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return entities.ErrConflict
	}

	return err
}
