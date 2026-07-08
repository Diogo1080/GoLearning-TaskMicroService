package store

import (
	"backendGo/internal/entities"
	"database/sql"
	"fmt"
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

func (r *SQLiteUserRepository) UpdateUser(userUpdates map[string]entities.User) (int64, error) {
	query := "UPDATE users SET "
	args := []interface{}{}
	i := 0

	//Loop through all of the changes and add them to the querry
	for key, value := range userUpdates {
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
	args = append(args, userUpdates["ID"])

	//Execute
	result, err := r.DB.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (r *SQLiteUserRepository) DeleteUser(id int) (int64, error) {

	result, err := r.DB.Exec("DELETE users WHERE id:?", id)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
