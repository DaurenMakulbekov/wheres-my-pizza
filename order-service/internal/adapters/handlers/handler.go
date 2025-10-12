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
	}
	var logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))

	result, err := hd.orderService.CreateOrder(order)
	if err != nil {
		if errors.Is(err, domain.ErrorBadRequest) {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			var m = map[string]string{"error": "Incorrect input"}
			encoder := json.NewEncoder(w)
			encoder.SetIndent("", " ")
			var err = encoder.Encode(m)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}

			logger.Error("Incorrect input", "method", "POST", "status", 400)
		} else if errors.Is(err, domain.InternalServerError) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			var m = map[string]string{"error": ""}
			encoder := json.NewEncoder(w)
			encoder.SetIndent("", " ")
			var err = encoder.Encode(m)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}

			logger.Error("", "method", "POST", "status", 500)
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

	logger.Debug("Order successfully created", "method", "POST", "status", 200)
}
