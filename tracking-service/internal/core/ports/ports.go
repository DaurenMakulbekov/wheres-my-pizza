package ports

import "wheres-my-pizza/tracking-service/internal/core/domain"

type PostRepo interface {
	GetOrder(string) (*domain.Order, error)
	GetOrderStatus(string) (*[]domain.Status, error)
	GetWorkersStatus() (*[]domain.Worker, error)
}

type PostService interface {
	GetOrder(string) (*domain.Order, error)
	GetOrderStatus(string) (*[]domain.Status, error)
	GetWorkersStatus() (*[]domain.Worker, error)
}
