package ports

import (
	"wheres-my-pizza/kitchen-worker/internal/core/domain"
)

type ConsumerService interface {
	Register(worker domain.Worker) (error, error)
}

type Database interface {
	Register(worker domain.Worker) error
}

type Consumer interface {
	RegisterConsumer() error
	Reconnect()
	IsClosed() (bool, bool)
	Close()
}
