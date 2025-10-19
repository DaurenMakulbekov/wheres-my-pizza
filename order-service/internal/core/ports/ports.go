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
