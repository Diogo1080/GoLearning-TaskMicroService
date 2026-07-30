package store

import (
	entities "backendGo/internal/domain"
	"database/sql"
)

type SQLiteAuthRepository struct {
	DB *sql.DB
}

func NewSQLiteAuthRepository(db *sql.DB) *SQLiteAuthRepository {
	return &SQLiteAuthRepository{DB: db}
}

func (r *SQLiteAuthRepository) GetUserByUsername(username string) (entities.User, error) {
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

	return entities.User{}, entities.ErrNotFound

}
