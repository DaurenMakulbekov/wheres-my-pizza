package ports

import (
	"wheres-my-pizza/kitchen-worker/internal/core/domain"
)

type ConsumerService interface {
	Register(workerName, orderTypes string, heartbeatInterval, prefetch int) error
	Start()
}

type Database interface {
	Register(worker domain.Worker) error
	GetWorker(name string) (domain.Worker, error)
	UpdateWorker(worker domain.Worker) error
	UpdateWorkerStatus(worker domain.Worker) error
	GetOrderStatus(order domain.Order) (string, error)
	UpdateOrder(worker domain.Worker, order domain.Order) error
	UpdateOrderReady(worker domain.Worker, order domain.Order) error
}

type Consumer interface {
	ReadMessages(orderTypes []string, prefetch int, out chan string, m map[string]chan bool)
	PublishStatusUpdate(message domain.Message) error
	Reconnect()
	IsClosed() (bool, bool)
	Close()
}
