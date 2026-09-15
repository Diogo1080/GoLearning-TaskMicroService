package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/identity"
	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/service"
	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/store"
	server "github.com/Diogo1080/GoLearning-TaskMicroService/internal/transport/http"
	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	db, err := store.Connect(store.GetConnectionURL())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	r := gin.Default()

	// Connect to Auth gRPC service
	identityAddr := os.Getenv("IDENTITY-SERVICE-ADDR")
	if identityAddr == "" {
		log.Fatal("IDENTITY-SERVICE-ADDR environment variable is not set")
		return
	}

	identityClient, err := identity.NewClient(identityAddr)
	if err != nil {
		log.Fatalf("Failed to connect to auth service: %v", err)
	}
	defer identityClient.Close()

	// Build auth middleware (uses gRPC validation)
	identityMiddleware := middleware.NewIdentityMiddlewareBuilder(identityClient).Build()

	// Initialize repositories
	taskRepo := store.NewSQLiteTaskRepository(db)

	// Initialize services (web service layer - no auth logic)
	taskService := service.NewTaskService(taskRepo)

	// Initialize handlers with auth client
	taskHandler := server.NewTaskHandler(taskService)

	//Register routes with auth middleware
	server.RegisterRoutes(r, taskHandler, identityMiddleware)

	address := fmt.Sprintf(":%s", os.Getenv("PORT"))
	log.Printf("Starting server on %s", address)
	if err := http.ListenAndServe(address, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
