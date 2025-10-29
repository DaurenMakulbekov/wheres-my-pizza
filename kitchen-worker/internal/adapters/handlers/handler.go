package handlers

import (
	"log/slog"
	"os"
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

func (hd *handler) RegisterHandler(workerName, orderTypes string, heartbeatInterval, prefetch int) {
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

	var err = hd.consumerService.Register(workerName, orderTypes, heartbeatInterval, prefetch)
	if err != nil {
		slog.Error("Worker registration failed", "service", "kitchen-worker", "hostname", "kitchen-worker", "request_id", "worker_registration", "action", "worker_registration_failed", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("Worker successfully registered", "service", "kitchen-worker", "hostname", "kitchen-worker", "request_id", "worker_registration", "action", "worker_registered")
}

func (hd *handler) ConsumerHandler() {
	var err = hd.consumerService.Start()
	if err != nil {
		slog.Error("Failed to get messages", "service", "kitchen-worker", "hostname", "kitchen-worker", "request_id", "message_processing", "action", "message_processing_failed", slog.Any("error", err))
	}
}
