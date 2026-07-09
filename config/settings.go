package config

import "os"

var (
	TODO_PORT   = getEnv("TODO_PORT", "8080")
	TODO_DBFILE = getEnv("TODO_DBFILE", "scheduler.db")
	TODO_PASS   = getEnv("TODO_PASS", "")
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
