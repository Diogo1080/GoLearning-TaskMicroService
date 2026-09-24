package tests

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/store"
	_ "github.com/lib/pq"
)

func openMigrationTestDB(t *testing.T) *sql.DB {
	t.Helper()

	databaseURL := os.Getenv("MIGRATION_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("MIGRATION_TEST_DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatalf("open migration test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("ping migration test database: %v", err)
	}

	return db
}

func resetMigrationDatabase(t *testing.T, db *sql.DB) {
	t.Helper()

	for _, statement := range []string{
		"DROP TABLE IF EXISTS tasks CASCADE",
		"DROP TABLE IF EXISTS schema_migrations",
	} {
		if _, err := db.ExecContext(context.Background(), statement); err != nil {
			t.Fatalf("reset migration database with %q: %v", statement, err)
		}
	}
}

func TestMigrationsFromEmptyDatabase(t *testing.T) {
	db := openMigrationTestDB(t)
	resetMigrationDatabase(t, db)
	t.Cleanup(func() { resetMigrationDatabase(t, db) })

	if err := store.RunMigrations(db); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	var version int
	var dirty bool
	if err := db.QueryRow("SELECT version, dirty FROM schema_migrations").Scan(&version, &dirty); err != nil {
		t.Fatalf("read migration version: %v", err)
	}
	if version != 2 || dirty {
		t.Fatalf("expected migration version 2 and clean state, got version %d dirty %t", version, dirty)
	}

	assertColumnExists(t, db, "due_date")
	assertIndexExists(t, db, "idx_tasks_user_id")
	assertIndexExists(t, db, "idx_tasks_user_due_date")
	assertConstraintExists(t, db, "tasks_priority_check")
}

func TestMigrationsFromPreviousSchemaPreserveData(t *testing.T) {
	db := openMigrationTestDB(t)
	resetMigrationDatabase(t, db)
	t.Cleanup(func() { resetMigrationDatabase(t, db) })

	_, err := db.Exec(`
		CREATE TABLE tasks (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			priority INTEGER NOT NULL,
			completed BOOLEAN NOT NULL,
			dueDate DATE
		)`)
	if err != nil {
		t.Fatalf("create previous schema: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO tasks (user_id, title, description, priority, completed, dueDate)
		VALUES (7, 'Existing task', 'Preserve this row', 2, false, '2026-10-01')`)
	if err != nil {
		t.Fatalf("insert previous-schema data: %v", err)
	}

	if err := store.RunMigrations(db); err != nil {
		t.Fatalf("apply migrations to previous schema: %v", err)
	}

	var title, description string
	var dueDate string
	if err := db.QueryRow("SELECT title, description, due_date::text FROM tasks WHERE user_id = 7").Scan(&title, &description, &dueDate); err != nil {
		t.Fatalf("read preserved task: %v", err)
	}
	if title != "Existing task" || description != "Preserve this row" || dueDate != "2026-10-01" {
		t.Fatalf("preserved task mismatch: %q, %q, %q", title, description, dueDate)
	}

	assertColumnExists(t, db, "due_date")
	assertIndexExists(t, db, "idx_tasks_user_id")
	assertIndexExists(t, db, "idx_tasks_user_due_date")
	assertConstraintExists(t, db, "tasks_priority_check")
}

func assertColumnExists(t *testing.T, db *sql.DB, column string) {
	t.Helper()

	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_name = 'tasks' AND column_name = $1
		)`, column).Scan(&exists)
	if err != nil {
		t.Fatalf("check column %q: %v", column, err)
	}
	if !exists {
		t.Fatalf("expected tasks.%s to exist", column)
	}
}

func assertIndexExists(t *testing.T, db *sql.DB, index string) {
	t.Helper()

	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = $1)", index).Scan(&exists)
	if err != nil {
		t.Fatalf("check index %q: %v", index, err)
	}
	if !exists {
		t.Fatalf("expected index %s to exist", index)
	}
}

func assertConstraintExists(t *testing.T, db *sql.DB, constraint string) {
	t.Helper()

	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM pg_constraint
			WHERE conname = $1
		)`, constraint).Scan(&exists)
	if err != nil {
		t.Fatalf("check constraint %q: %v", constraint, err)
	}
	if !exists {
		t.Fatalf("expected constraint %s to exist", constraint)
	}
}
