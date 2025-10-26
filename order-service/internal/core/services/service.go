package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
	"wheres-my-pizza/order-service/internal/core/domain"
	"wheres-my-pizza/order-service/internal/core/ports"
)

type service struct {
	orderRepository ports.OrderRepository
	publisher       ports.Publisher
	ctx             context.Context
	ctxCansel       context.CancelFunc
}

func NewOrderService(orderRepo ports.OrderRepository, publisherRepo ports.Publisher, ctxMain context.Context) *service {
	ctx, ctxCansel := context.WithCancel(ctxMain)

	return &service{
		orderRepository: orderRepo,
		publisher:       publisherRepo,
		ctx:             ctx,
		ctxCansel:       ctxCansel,
	}
}

func CheckCustomerName(name string) error {
	if len(name) < 1 || len(name) > 100 {
		return fmt.Errorf("Incorrect customer name. Must be 1 - 100 characters.")
	}

	for i := range name {
		if name[i] >= 65 && name[i] <= 90 || name[i] >= 97 && name[i] <= 122 {
			continue
		} else if name[i] == 32 || name[i] == 45 || name[i] == 34 || name[i] == 39 {
			continue
		} else {
			return fmt.Errorf("Incorrect customer name. Must not contain special characters other than spaces, hyphens and apostrophes.")
		}
	}

	return nil
}

func CheckOrderItems(orderItems []domain.OrderItem) error {
	if len(orderItems) < 1 || len(orderItems) > 20 {
		return fmt.Errorf("Incorrect order items. Must contain between 1 and 20 items.")
	}

	for i := range orderItems {
		if len(orderItems[i].Name) < 1 || len(orderItems[i].Name) > 50 {
			return fmt.Errorf("Incorrect name. Must be 1 - 50 characters.")
		} else if orderItems[i].Quantity < 1 || orderItems[i].Quantity > 10 {
			return fmt.Errorf("Incorrect quantity. Must be between 1 and 10.")
		} else if orderItems[i].Price < 0.01 || orderItems[i].Price > 999.99 {
			return fmt.Errorf("Incorrect price. Must be between 0.01 and 999.99.")
		}
	}

	return nil
}

func CheckInput(order domain.Order) error {
	var err = CheckCustomerName(order.CustomerName)
	if err != nil {
		return err
	}

	if !slices.Contains([]string{"dine_in", "takeout", "delivery"}, order.Type) {
		return fmt.Errorf("Incorrect order type. Must be one of: 'dine_in', 'takeout', 'delivery'")
	}

	if order.Type == "dine_in" {
		if order.TableNumber < 1 || order.TableNumber > 100 {
			return fmt.Errorf("Incorrect input for order type 'dine_in'. Table number must be between 1 and 100.")
		}
		if len(order.DeliveryAddress) > 0 {
			return fmt.Errorf("Incorrect input for order type 'dine_in'. Delivery address must not be present.")
		}
	} else if order.Type == "delivery" {
		if len(order.DeliveryAddress) < 10 {
			return fmt.Errorf("Incorrect delivery address. Must be min 10 characters.")
		}
		if order.TableNumber > 0 {
			return fmt.Errorf("Incorrect input for order type 'delivery'. Table number must not be present.")
		}
	}

	if err := CheckOrderItems(order.Items); err != nil {
		return err
	}

	return nil
}

func GetTotalAmount(orderItems []domain.OrderItem) float64 {
	var totalAmount float64

	for i := range orderItems {
		totalAmount += orderItems[i].Price * float64(orderItems[i].Quantity)
	}

	return totalAmount
}

func (service *service) CreateOrderNumber() string {
	var timeNow = time.Now()
	var value = fmt.Sprintf("%d%d%d", timeNow.Year(), timeNow.Month(), timeNow.Day())

	number, err := service.orderRepository.GetOrderNumber()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}

	if len(number) == 0 {
		number = "ORD_" + value + "_" + "001"
	} else {
		var result = strings.Split(number, "_")

		if result[1] != value {
			number = "ORD_" + value + "_" + "001"
		} else {
			var length = len(number)
			var length_n = len(result[2])

			i, err := strconv.Atoi(number[length-length_n:])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			i += 1
			s := strconv.Itoa(i)
			var length1 = len(s)
			if length1 < 3 {
				number = number[:length-length1] + s
			} else {
				number = number[:length-length_n] + s
			}
		}
	}

	return number
}

func SetOrderItemsCreatedAt(orderItems []domain.OrderItem) {
	for i := range orderItems {
		orderItems[i].CreatedAt = time.Now().UTC()
	}
}

func (service *service) Push(message domain.Order) {
	go func() {
		service.publisher.Reconnect()
		for {
			select {
			case <-time.After(5 * time.Second):
				var err = service.publisher.Publish(message)
				if err != nil {
					continue
				} else {
					log.Println("Message published to RabbitMQ")

					return
				}
			case <-service.ctx.Done():
				return
			}
		}
	}()
}

func (service *service) CreateOrder(order domain.Order) (domain.Result, error, error) {
	var result domain.Result

	var err = CheckInput(order)
	if err != nil {
		return result, domain.ErrorBadRequest, err
	}

	order.TotalAmount = GetTotalAmount(order.Items)

	if order.TotalAmount > 100 {
		order.Priority = 10
	} else if order.TotalAmount >= 50 && order.TotalAmount <= 100 {
		order.Priority = 5
	} else {
		order.Priority = 1
	}

	order.Number = service.CreateOrderNumber()
	order.Status = "received"
	order.CreatedAt = time.Now().UTC()
	order.UpdatedAt = time.Now().UTC()
	SetOrderItemsCreatedAt(order.Items)

	err = service.orderRepository.CreateOrder(order)
	if err != nil {
		return result, domain.InternalServerError, err
	}

	err = service.publisher.Publish(order)
	if err != nil {
		var chanClosed, connClosed = service.publisher.IsClosed()
		if chanClosed && connClosed {
			service.Push(order)
		}

		return result, domain.InternalServerError, err
	}
	log.Println("Message published to RabbitMQ")

	result = domain.Result{
		OrderNumber: order.Number,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
	}

	return result, nil, nil
}

func (service *service) Close() {
	service.ctxCansel()
	service.publisher.Close()
}
