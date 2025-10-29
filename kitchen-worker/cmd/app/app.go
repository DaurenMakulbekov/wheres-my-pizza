package app

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"
	"wheres-my-pizza/kitchen-worker/internal/adapters/handlers"
	"wheres-my-pizza/kitchen-worker/internal/adapters/repositories/postgres"
	"wheres-my-pizza/kitchen-worker/internal/adapters/repositories/rabbitmq"
	"wheres-my-pizza/kitchen-worker/internal/core/services"
	"wheres-my-pizza/kitchen-worker/internal/infrastructure/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Run(workerName, orderTypes string, heartbeatInterval, prefetch int) {
	config := config.NewAppConfig()
	ctx := context.Background()

	db, err := postgres.ConnectDB(config.DB)
	if err != nil {
		log.Fatal("Unable to connect to database")
	}

	var database = postgres.NewDatabaseRepository(db)
	var consumer = rabbitmq.NewRabbitMQRepository(config.RabbitMQ, ctx)
	var consumerService = services.NewConsumerService(database, consumer, ctx)
	var handler = handlers.NewConsumerHandler(consumerService)

	handler.RegisterHandler(workerName, orderTypes, heartbeatInterval, prefetch)
	handler.ConsumerHandler()

	signalCtx, signalCtxStop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGTSTP)
	defer signalCtxStop()

	<-signalCtx.Done()

	log.Println("Shutting down process...")
	consumerService.Close()
	time.Sleep(5 * time.Second)

	log.Println("Process shutdown complete")
}
