package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"wheres-my-pizza/order-service/internal/core/domain"
	"wheres-my-pizza/order-service/internal/core/ports"
)

type handler struct {
	orderRepository ports.OrderRepository
	orderService    ports.OrderService
}

func NewOrderHandler(repository ports.OrderRepository, service ports.OrderService) *handler {
	return &handler{
		orderRepository: repository,
		orderService:    service,
	}
}

func (hd *handler) CreateOrderHandler(w http.ResponseWriter, req *http.Request) {
	var decoder = json.NewDecoder(req.Body)
	var order domain.Order

	var err = decoder.Decode(&order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var opts = &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Key = "timestamp"
			}
			if a.Key == slog.MessageKey {
				a.Key = "message"
			}

			return a
		},
	}
	var logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	result, err, message := hd.orderService.CreateOrder(order)
	if err != nil {
		if errors.Is(err, domain.ErrorBadRequest) {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			var m = map[string]string{"error": message.Error()}
			encoder := json.NewEncoder(w)
			encoder.SetIndent("", " ")
			var err_ = encoder.Encode(m)
			if err_ != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err_)
			}

			slog.Error("Incorrect input", "service", "order-service", "hostname", "order-service", "request_id", "create_order", "action", "validation_failed", slog.Any("error", err))
		} else if errors.Is(err, domain.InternalServerError) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			var m = map[string]string{"error": message.Error()}
			encoder := json.NewEncoder(w)
			encoder.SetIndent("", " ")
			var err_ = encoder.Encode(m)
			if err_ != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err_)
			}

			slog.Error("Failed to create order", "service", "order-service", "hostname", "order-service", "request_id", "create_order", "action", "db_transaction_failed", slog.Any("error", err))
		}

		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", " ")
	err = encoder.Encode(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}

	slog.Debug("Order successfully created", "service", "order-service", "hostname", "order-service", "request_id", "create_order", "action", "order_received")
}
