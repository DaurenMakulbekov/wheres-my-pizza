package app

import (
	"log"
	"net/http"
	"wheres-my-pizza/tracking-service/internal/adapters/repositories"
	"wheres-my-pizza/tracking-service/internal/config"
	"wheres-my-pizza/tracking-service/internal/core/services"
	"wheres-my-pizza/tracking-service/internal/handlers/posthandler"
	"wheres-my-pizza/tracking-service/internal/handlers/routes"
)

func Run() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	userRepo, err := repositories.NewUserRepository(cfg)
	if err != nil {
		log.Fatalf("failed to init user repository %v", err)

	}

	svc := services.NewService(userRepo)

	handlers := posthandler.NewHandler(svc)

	mux := http.NewServeMux()
	routes.SetupRoutes(mux, handlers)
	if err := http.ListenAndServe(":3002", mux); err != nil {
		log.Fatal("Server error:", err)
	}
}
