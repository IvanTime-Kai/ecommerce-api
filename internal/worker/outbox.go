package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/Ivantime-Kai/ecommerce-api/internal/kafka"
	"github.com/Ivantime-Kai/ecommerce-api/internal/repository"
)

type OutboxWorker struct {
	repository repository.Querier
	kafka      kafka.KafkaProducer
	interval   time.Duration
}

func NewOutboxWorker(repository repository.Querier, kafka kafka.KafkaProducer, interval time.Duration) *OutboxWorker {
	return &OutboxWorker{
		repository: repository,
		kafka:      kafka,
		interval:   interval,
	}
}

func (w *OutboxWorker) process(ctx context.Context) {
	events, err := w.repository.GetPendingOutboxEvents(ctx)

	if err != nil {
		slog.Error("failed to get pending outbox events", "error", err)
		return
	}

	for _, event := range events {
		payload := event.Payload

		if err := w.kafka.Publish(ctx, []byte(event.ID.String()), payload); err != nil {
			slog.Error("failed to publish outbox event", "id", event.ID, "error", err)
			continue
		}

		if err := w.repository.MarkOutboxEventProcessed(ctx, event.ID); err != nil {
			slog.Error("failed to mark outbox event processed", "id", event.ID, "error", err)
		}
	}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.process(ctx)
		}
	}
}
