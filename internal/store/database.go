package store

import (
	"database/sql"
	"log"

	"backendGo/config"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() *sql.DB {
	db, err := sql.Open("sqlite3", config.TODO_DBFILE)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	_, err = db.Exec(`DROP TABLE IF EXISTS tasks`)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		priority INTEGER NOT NULL,
		completed BOOLEAN NOT NULL,
		dueDate DATETIME
	)`)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
	)`)

	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	return db
}
