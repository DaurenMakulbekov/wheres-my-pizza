package services

import (
	"context"
	"time"
	"wheres-my-pizza/kitchen-worker/internal/core/domain"
	"wheres-my-pizza/kitchen-worker/internal/core/ports"
)

type service struct {
	database  ports.Database
	consumer  ports.Consumer
	ctx       context.Context
	ctxCansel context.CancelFunc
}

func NewConsumerService(database ports.Database, consumerRepo ports.Consumer, ctxMain context.Context) *service {
	ctx, ctxCansel := context.WithCancel(ctxMain)

	return &service{
		database:  database,
		consumer:  consumerRepo,
		ctx:       ctx,
		ctxCansel: ctxCansel,
	}
}

func (service *service) Push(message domain.Order) {
	go func() {
		service.consumer.Reconnect()
		for {
			select {
			case <-time.After(5 * time.Second):
				//var err = service.consumer.Publish(message)
				//if err != nil {
				//	continue
				//} else {
				//	log.Println("Message published to RabbitMQ")

				//	return
				//}
			case <-service.ctx.Done():
				return
			}
		}
	}()
}

func (service *service) Register(worker domain.Worker) (error, error) {
	var err = service.database.Register(worker)
	if err != nil {
		return domain.ErrorBadRequest, err
	}

	return nil, nil
}

func (service *service) Close() {
	service.ctxCansel()
	service.consumer.Close()
}
