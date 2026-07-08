package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

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
	UserRepo := store.NewSQLiteUserRepository(db)

	taskService := service.NewTaskService(taskRepo)
	authService := service.NewAuthService(UserRepo)

	router := routes.RegisterRoutes(taskService, authService, rds)

	fileServer := http.FileServer(http.Dir("./web"))
	router.GET("/*", func(c *gin.Context) {
		if filepath.Ext(c.Request.URL.Path) == ".css" {
			c.Header("Content-Type", "text/css")
		}
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	address := fmt.Sprintf(":%s", os.Getenv("TODO_PORT"))
	log.Printf("Starting server on %s", address)
	if err := http.ListenAndServe(address, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
