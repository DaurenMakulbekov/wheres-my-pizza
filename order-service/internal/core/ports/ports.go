package ports

import (
	"wheres-my-pizza/order-service/internal/core/domain"
)

type OrderService interface {
	CreateOrder(order domain.Order) (domain.Result, error, error)
}

type OrderRepository interface {
	CreateOrder(order domain.Order) error
	GetOrderNumber() (string, error)
}

type Publisher interface {
	Publish(order domain.Order) error
	Reconnect()
	IsClosed() (bool, bool)
	Close()
}
