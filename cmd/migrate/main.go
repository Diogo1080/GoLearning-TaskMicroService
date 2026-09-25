package main

import (
	"log"

	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/config"
	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/store"
)

func main() {
	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	db, err := store.Connect(cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := store.RunMigrations(db); err != nil {
		log.Fatalf("Failed to apply database migrations: %v", err)
	}
}
