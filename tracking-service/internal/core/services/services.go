package services

import (
	"wheres-my-pizza/tracking-service/internal/core/domain"
	"wheres-my-pizza/tracking-service/internal/core/ports"
)

type service struct {
	Repo ports.PostRepo
}

func NewService(repo ports.PostRepo) *service {
	return &service{Repo: repo}
}

func (s *service) GetOrder(orderNum string) (*domain.Order, error) {

	return s.Repo.GetOrder(orderNum)
}

func (s *service) GetOrderStatus(orderNum string) (*[]domain.Status, error) {
	return s.Repo.GetOrderStatus(orderNum)

}
func (s *service) GetWorkersStatus() (*[]domain.Worker, error) {
	return s.Repo.GetWorkersStatus()
}
