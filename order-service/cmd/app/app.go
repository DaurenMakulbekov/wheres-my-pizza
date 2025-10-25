package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
	"wheres-my-pizza/order-service/internal/adapters/handlers"
	"wheres-my-pizza/order-service/internal/adapters/repositories/postgres"
	"wheres-my-pizza/order-service/internal/adapters/repositories/rabbitmq"
	"wheres-my-pizza/order-service/internal/core/services"
	"wheres-my-pizza/order-service/internal/infrastructure/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Run(port string, maxConcurrent int) {
	config := config.NewAppConfig()
	ctx := context.Background()

	db, err := postgres.ConnectDB(config.DB)
	if err != nil {
		log.Fatal("Unable to connect to database")
	}

	var orderRepository = postgres.NewOrderRepository(db)
	var publisher = rabbitmq.NewRabbitMQRepository(config.RabbitMQ)
	var orderService = services.NewOrderService(orderRepository, publisher)
	var handler = handlers.NewOrderHandler(orderRepository, orderService)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler.Routes(),
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
