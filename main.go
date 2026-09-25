package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/config"
	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/identity"
	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/observability"
	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/service"
	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/store"
	server "github.com/Diogo1080/GoLearning-TaskMicroService/internal/transport/http"
	"github.com/Diogo1080/GoLearning-TaskMicroService/internal/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	cfg, err := config.Load("../.env")
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

	metrics := observability.NewMetrics(prometheus.DefaultRegisterer, "task-service", "dev", cfg.AppEnv)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), otelgin.Middleware("task-service"), middleware.PrometheusMetrics(metrics))

	// Connect to Auth gRPC service
	identityClient, err := identity.NewClient(cfg.IdentityServiceAddr, metrics)
	if err != nil {
		log.Fatalf("Failed to connect to auth service: %v", err)
	}
	defer identityClient.Close()
	identityReadyContext, identityReadyCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer identityReadyCancel()
	if err := identityClient.WaitReady(identityReadyContext); err != nil {
		log.Fatalf("Auth service is not ready: %v", err)
	}

	// Build auth middleware (uses gRPC validation)
	identityMiddleware := middleware.NewIdentityMiddlewareBuilder(identityClient, cfg.AppEnv).Build()

	// Initialize repositories
	taskRepo := store.NewSQLiteTaskRepository(db)

	// Initialize services (web service layer - no auth logic)
	taskService := service.NewTaskService(taskRepo)

	// Initialize handlers with auth client
	taskHandler := server.NewTaskHandler(taskService)

	//Register routes with auth middleware
	server.RegisterRoutes(r, taskHandler, identityMiddleware,
		db.PingContext,
		identityClient.Ready,
	)

	address := fmt.Sprintf(":%s", cfg.Port)
	httpServer := &http.Server{
		Addr:              address,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	shutdownContext, stopSignals := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopSignals()

	log.Printf("Starting task-service on %s (environment=%s, version=%s)", address, cfg.AppEnv, "dev")
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	case <-shutdownContext.Done():
		log.Printf("Shutting down server: %v", shutdownContext.Err())
		shutdownDeadline, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownDeadline); err != nil {
			log.Printf("HTTP server shutdown failed: %v", err)
		}
	}
}
