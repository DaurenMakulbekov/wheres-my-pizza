package services

import (
	"wheres-my-pizza/order-service/internal/core/domain"
	"wheres-my-pizza/order-service/internal/core/ports"
)

type service struct {
	orderRepository ports.OrderRepository
}

func NewOrderService(repository ports.OrderRepository) *service {
	return &service{
		orderRepository: repository,
	}
}

func (service *service) CreateOrder(order domain.Order) (domain.Result, error) {
	var result domain.Result

	var err = service.orderRepository.CreateOrder(order)
	if err != nil {
		return result, domain.InternalServerError
	}

	result = domain.Result{
		OrderNumber: order.Number,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
	}

	return result, err
}
