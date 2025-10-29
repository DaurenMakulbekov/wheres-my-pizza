package ports

import (
	"wheres-my-pizza/kitchen-worker/internal/core/domain"
)

type ConsumerService interface {
	Register(workerName, orderTypes string, heartbeatInterval, prefetch int) error
	Start() error
}

type Database interface {
	Register(worker domain.Worker) error
	GetWorker(name string) (domain.Worker, error)
	UpdateWorker(worker domain.Worker) error
}

type Consumer interface {
	ReadMessages(orderType string, prefetch int, out chan string, done chan bool) error
	Reconnect()
	IsClosed() (bool, bool)
	Close()
}
