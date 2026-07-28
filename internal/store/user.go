package store

import (
	entities "backendGo/internal/domain"
	"database/sql"
)

type SQLiteUserRepository struct {
	DB *sql.DB
}

func NewSQLiteUserRepository(db *sql.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{DB: db}
}

func (r *SQLiteUserRepository) CreateUser(user entities.User) (entities.User, error) {
	result, err := r.DB.Exec("INSERT INTO users (username, password) VALUES (?, ?)", user.Username, user.Password)
	if err != nil {
		return entities.User{}, err
	}

	user.ID, _ = result.LastInsertId()
	return user, nil
}

func (r *SQLiteUserRepository) GetUserByUsername(username string) (entities.User, error) {
	result, err := r.DB.Query("SELECT id, username, password FROM users WHERE username = ?", username)
	if err != nil {
		return entities.User{}, err
	}

	defer result.Close()

	if result.Next() {
		var user entities.User
		err := result.Scan(&user.ID, &user.Username, &user.Password)
		if err != nil {
			return entities.User{}, err
		}
		return user, nil
	}

	return entities.User{}, sql.ErrNoRows
}

func (r *SQLiteUserRepository) GetUserByID(id int) (entities.User, error) {
	result, err := r.DB.Query("SELECT id, username, password FROM users WHERE id = ?", id)
	if err != nil {
		return entities.User{}, err
	}

	defer result.Close()

	if result.Next() {
		var user entities.User
		err := result.Scan(&user.ID, &user.Username, &user.Password)
		if err != nil {
			return entities.User{}, err
		}
		return user, nil
	}

	return entities.User{}, sql.ErrNoRows
}

func (r *SQLiteUserRepository) UpdateUser(user entities.User, id int) (entities.User, error) {
	result, err := r.DB.Exec("UPDATE users SET password = ? WHERE id = ?",
		user.Password, id)
	if err != nil {
		return entities.User{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return entities.User{}, err
	}

	if rowsAffected == 0 {
		return entities.User{}, entities.ErrNotFound
	}

	return user, nil
}

func (r *SQLiteUserRepository) DeleteUser(id int) error {

	_, err := r.DB.Exec("DELETE FROM users WHERE id:?", id)

	if err != nil {
		return err
	}

	return nil
}
