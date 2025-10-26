package handlers

import (
	"errors"
	"log/slog"
	"os"
	"wheres-my-pizza/kitchen-worker/internal/core/domain"
	"wheres-my-pizza/kitchen-worker/internal/core/ports"
)

type handler struct {
	consumerService ports.ConsumerService
}

func NewConsumerHandler(service ports.ConsumerService) *handler {
	return &handler{
		consumerService: service,
	}
}

func (hd *handler) WorkerHandler(workerName, orderTypes string, heartbeatInterval, prefetch int) {
	var worker domain.Worker

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

	err, _ := hd.consumerService.Register(worker)
	if err != nil {
		if errors.Is(err, domain.ErrorBadRequest) {

			slog.Error("Worker registration failed", "service", "kitchen-worker", "hostname", "kitchen-worker", "request_id", "worker_registration", "action", "worker_registration_failed", slog.Any("error", err))
		} else if errors.Is(err, domain.InternalServerError) {

			slog.Error("Failed to register worker", "service", "kitchen-worker", "hostname", "kitchen-worker", "request_id", "worker_registration", "action", "db_connection_failed", slog.Any("error", err))
		}

		return
	}

	slog.Info("Worker successfully registered", "service", "kitchen-worker", "hostname", "kitchen-worker", "request_id", "worker_registration", "action", "worker_registered")
}
