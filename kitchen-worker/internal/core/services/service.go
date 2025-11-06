package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"slices"
	"strings"
	"time"

	"wheres-my-pizza/kitchen-worker/internal/core/domain"
	"wheres-my-pizza/kitchen-worker/internal/core/ports"
)

type service struct {
	database          ports.Database
	consumer          ports.Consumer
	ctx               context.Context
	ctxCansel         context.CancelFunc
	worker            domain.Worker
	orderTypes        []string
	heartbeatInterval int
	prefetch          int
}

func NewConsumerService(database ports.Database, consumerRepo ports.Consumer, ctxMain context.Context) *service {
	ctx, ctxCansel := context.WithCancel(ctxMain)
	var orderTypes = []string{"dine_in", "takeout", "delivery"}

	return &service{
		database:   database,
		consumer:   consumerRepo,
		ctx:        ctx,
		ctxCansel:  ctxCansel,
		orderTypes: orderTypes,
	}
}

func CheckWorkerName(name string) error {
	if len(name) < 1 || len(name) > 100 {
		return fmt.Errorf("Incorrect worker name. Must be 1 - 100 characters.")
	}

	for i := range name {
		if name[i] >= 65 && name[i] <= 90 || name[i] >= 97 && name[i] <= 122 {
			continue
		} else if name[i] == 32 || name[i] == 45 || name[i] == 34 || name[i] == 39 {
			continue
		} else {
			return fmt.Errorf("Incorrect worker name. Must not contain special characters other than spaces, hyphens and apostrophes.")
		}
	}

	return nil
}

func GetOrderTypes(orderTypes string) []string {
	var result1 = strings.Split(orderTypes, ",")
	var result []string

	for i := range result1 {
		var value = strings.Trim(result1[i], " ")
		result = append(result, value)
	}

	return result
}

func CheckOrderTypes(orderTypes []string) error {
	for i := range orderTypes {
		if !slices.Contains([]string{"dine_in", "takeout", "delivery"}, orderTypes[i]) {
			return fmt.Errorf("Incorrect order type. Must be one of: 'dine_in', 'takeout', 'delivery'")
		}
	}

	return nil
}

func (service *service) Register(workerName, orderTypes string, heartbeatInterval, prefetch int) error {
	if err := CheckWorkerName(workerName); err != nil {
		return err
	}

	if len(orderTypes) > 0 {
		var result = GetOrderTypes(orderTypes)

		var err = CheckOrderTypes(result)
		if err != nil {
			return err
		}

		service.orderTypes = result
	}

	if heartbeatInterval < 1 || heartbeatInterval > 600 {
		return fmt.Errorf("heartbeat-interval must be between 1 and 600 seconds")
	}
	if prefetch < 1 || prefetch > 5 {
		return fmt.Errorf("prefetch count must be between 1 and 5")
	}

	var worker = domain.Worker{
		Name: workerName,
		Type: orderTypes,
	}

	user, err := service.database.GetWorker(workerName)
	if err != nil {
		err = service.database.Register(worker)
		if err != nil {
			return err
		}
	}

	if user.Status == "online" {
		return fmt.Errorf("Worker is already online")
	} else {
		worker.Status = "online"
		var err = service.database.UpdateWorker(worker)
		if err != nil {
			return err
		}
	}

	service.worker = worker
	service.heartbeatInterval = heartbeatInterval
	service.prefetch = prefetch

	return nil
}

func (service *service) CreateMessage(order domain.Order, oldStatus, newStatus string) domain.Message {
	var message = domain.Message{
		OrderNumber: order.Number,
		OldStatus:   oldStatus,
		NewStatus:   newStatus,
		ChangedBy:   service.worker.Name,
		Timestamp:   time.Now(),
	}

	if order.Type == "dine_in" {
		message.EstimatedCompletion = time.Now().Add(8 * time.Second)
	} else if order.Type == "takeout" {
		message.EstimatedCompletion = time.Now().Add(10 * time.Second)
	} else if order.Type == "delivery" {
		message.EstimatedCompletion = time.Now().Add(12 * time.Second)
	}

	return message
}

func (service *service) UpdateWorker(ticker *time.Ticker) {
	go func() {
		for {
			select {
			case <-service.ctx.Done():
				return
			case <-ticker.C:
				var err = service.database.UpdateWorkerStatus(service.worker)
				if err != nil {
					slog.Error("Failed to update worker status", "service", "kitchen-worker", "hostname", "kitchen-worker", "request_id", "heartbeat_sending", "action", "heartbeat_sent_failed", slog.Any("error", err))
				} else {
					slog.Debug("A heartbeat is successfully sent", "service", "kitchen-worker", "hostname", "kitchen-worker", "request_id", "heartbeat_sending", "action", "heartbeat_sent")
				}
			}
		}
	}()
}

func (service *service) Start() {
	var orderTypes = service.orderTypes
	var out = make(chan string)
	defer close(out)

	var m = make(map[string]chan bool)

	var ticker = time.NewTicker(time.Duration(service.heartbeatInterval) * time.Second)
	defer ticker.Stop()
	service.UpdateWorker(ticker)

	go service.consumer.ReadMessages(orderTypes, service.prefetch, out, m)

	for {
		select {
		case <-service.ctx.Done():
			return
		case message := <-out:
			var order domain.Order
			decoder := json.NewDecoder(strings.NewReader(message))

			err := decoder.Decode(&order)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error decode:", err)
			}

			if !slices.Contains(orderTypes, order.Type) {
				m[order.Type] <- true
				continue
			}

			err = service.database.UpdateOrder(service.worker, order)
			if err != nil {
				m[order.Type] <- true
				continue
			}

			var msg = service.CreateMessage(order, "received", "cooking")

			err = service.consumer.PublishStatusUpdate(msg)
			if err != nil {
				m[order.Type] <- true
				break
			}

			var pause time.Duration

			if order.Type == "dine_in" {
				pause = 8
			} else if order.Type == "takeout" {
				pause = 10
			} else if order.Type == "delivery" {
				pause = 12
			}

			select {
			case <-service.ctx.Done():
				msg = service.CreateMessage(order, "cooking", "ready")

				err = service.consumer.PublishStatusUpdate(msg)
				if err != nil {
					m[order.Type] <- true
					return
				}

				err = service.database.UpdateOrderReady(service.worker, order)
				if err != nil {
					m[order.Type] <- true
					return
				}

				m[order.Type] <- false

				return
			case <-time.After(pause * time.Second):
				msg = service.CreateMessage(order, "cooking", "ready")

				err = service.consumer.PublishStatusUpdate(msg)
				if err != nil {
					m[order.Type] <- true
					break
				}

				err = service.database.UpdateOrderReady(service.worker, order)
				if err != nil {
					m[order.Type] <- true
					continue
				}

				m[order.Type] <- false
			}
		}
	}
}

func (service *service) Close() {
	service.worker.Status = "offline"

	var err = service.database.UpdateWorker(service.worker)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	service.ctxCansel()
	service.consumer.Close()
}
