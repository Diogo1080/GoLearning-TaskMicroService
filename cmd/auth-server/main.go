package main

import (
	"log"
	"net"
	"os"

	authv1 "backendGo/api/auth/v1"
	service "backendGo/internal/service"
	store "backendGo/internal/store"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	rds := store.NewRedis()
	defer rds.Client.Close()

	repo := store.NewSQLiteAuthRepository(db)

	lis, err := net.Listen("tcp", ":"+os.Getenv("AUTH_PORT"))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	authv1.RegisterAuthServiceServer(grpcServer, service.NewAuthService(repo, rds))

	log.Printf("Auth gRPC server listening on port %s", os.Getenv("AUTH_PORT"))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
