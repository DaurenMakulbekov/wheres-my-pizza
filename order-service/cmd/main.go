package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
	"wheres-my-pizza/order-service/internal/adapters/handlers"
	"wheres-my-pizza/order-service/internal/adapters/repositories/postgres/repository"
	"wheres-my-pizza/order-service/internal/core/services"
	"wheres-my-pizza/order-service/internal/infrastructure/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	config := config.NewAppConfig()
	ctx := context.Background()

	var orderRepository = postgres.NewPostgresRepository(config.DB)
	var orderService = services.NewOrderService(orderRepository)
	var handler = handlers.NewOrderHandler(orderRepository, orderService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /orders", handler.CreateOrderHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	signalCtx, signalCtxStop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGTSTP)
	defer signalCtxStop()

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Listen and serve returned error: %v", err)
		}
	}()

	<-signalCtx.Done()

	log.Println("Shutting down server...")
	time.Sleep(5 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Error during shutdown: %v\n", err)
	}

	log.Println("Server shutdown complete")
}
