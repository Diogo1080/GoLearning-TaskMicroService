package store

import (
	"backendGo/internal/entities"
	"database/sql"
)

type SQLiteUserRepository struct {
	DB *sql.DB
}

func NewSQLiteUserRepository(db *sql.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{DB: db}
}

func (r *SQLiteUserRepository) CreateUser(user entities.User) (entities.User, error) {
	// Here you would implement the logic to insert the user into the SQLite database.
	// For demonstration purposes, we'll just return the user as is.
	return user, nil
}

func (r *SQLiteUserRepository) GetUserByUsername(username string) (entities.User, error) {
	// Here you would implement the logic to retrieve a user by their username from the SQLite database.
	// For demonstration purposes, we'll just return a default user and no error.
	return entities.User{}, nil
}

func (r *SQLiteUserRepository) UpdateUser(user entities.User) (entities.User, error) {
	// Here you would implement the logic to update a user in the SQLite database.
	// For demonstration purposes, we'll just return the user as is.
	return user, nil
}

func (r *SQLiteUserRepository) DeleteUser(id int) error {
	// Here you would implement the logic to delete a user by their ID from the SQLite database.
	// For demonstration purposes, we'll just return no error.
	return nil
}
