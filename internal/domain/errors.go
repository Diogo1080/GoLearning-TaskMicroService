package domain

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")

	ErrBadData = errors.New("bad data")

	ErrUnauthorized = errors.New("unauthorized")

	ErrDatabaseFailed = errors.New("Database failed")
)
