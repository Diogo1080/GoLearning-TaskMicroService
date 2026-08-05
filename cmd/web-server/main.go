package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"backendGo/internal/auth"
	"backendGo/internal/service"
	"backendGo/internal/store"
	server "backendGo/internal/transport/http"
	"backendGo/internal/transport/http/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println(err.Error())
	}

	db, err := store.Connect(store.GetConnectionURL())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{os.Getenv("FRONTEND_ORIGIN")},
		AllowMethods:     []string{"GET", "POST", "PATCH", "PUT", "DELETE"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Connect to Auth gRPC service
	authAddr := os.Getenv("AUTH_SERVICE_ADDR")
	if authAddr == "" {
		log.Fatal("AUTH_SERVICE_ADDR environment variable is not set")
		return
	}

	authClient, err := auth.NewClient(authAddr)
	if err != nil {
		log.Fatalf("Failed to connect to auth service: %v", err)
	}
	defer authClient.Close()

	// Build auth middleware (uses gRPC validation)
	authMiddleware := middleware.NewAuthMiddlewareBuilder(authClient).Build()

	// Initialize repositories
	taskRepo := store.NewSQLiteTaskRepository(db)
	userRepo := store.NewSQLiteUserRepository(db)

	// Initialize services (web service layer - no auth logic)
	taskService := service.NewTaskService(taskRepo)
	userService := service.NewUserService(userRepo)

	// Initialize handlers with auth client
	taskHandler := server.NewTaskHandler(taskService)
	userHandler := server.NewUserHandler(userService, authClient)
	authHandler := server.NewAuthHandler(authClient)

	//Register routes with auth middleware
	server.RegisterRoutes(r, taskHandler, userHandler, authHandler, authMiddleware)

	address := fmt.Sprintf(":%s", os.Getenv("PORT"))
	log.Printf("Starting server on %s", address)
	if err := http.ListenAndServe(address, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
