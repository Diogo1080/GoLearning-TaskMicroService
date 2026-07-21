package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"backendGo/internal/server"
	"backendGo/internal/service"
	"backendGo/internal/store"
	"backendGo/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	db := store.InitDB()
	defer db.Close()

	_ = godotenv.Load()

	for _, k := range []string{"ACCESS_SECRET", "REFRESH_SECRET"} {
		if os.Getenv(k) == "" {
			log.Fatalf("%s not set", k)
		}
	}

	rds := store.NewRedis()
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{os.Getenv("FRONTEND_ORIGIN")},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	taskRepo := store.NewSQLiteTaskRepository(db)
	userRepo := store.NewSQLiteUserRepository(db)
	authRepo := store.NewSQLiteAuthRepository(db)

	taskHandler := server.NewTaskHandler(service.NewTaskService(taskRepo))
	userHandler := server.NewUserHandler(service.NewUserService(userRepo))
	authHandler := server.NewAuthHandler(service.NewAuthService(authRepo), rds)

	routes.RegisterRoutes(r, taskHandler, userHandler, authHandler, rds)

	address := fmt.Sprintf(":%s", os.Getenv("PORT"))
	log.Printf("Starting server on %s", address)
	if err := http.ListenAndServe(address, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
