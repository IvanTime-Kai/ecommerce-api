package worker

import (
	"context"
	"log/slog"

	"github.com/Ivantime-Kai/ecommerce-api/internal/kafka"
	"github.com/Ivantime-Kai/ecommerce-api/internal/repository"
	"github.com/jackc/pgx/v5"
)

type OutboxWorker struct {
	dbURL      string
	repository repository.Querier
	kafka      kafka.KafkaProducer
}

func NewOutboxWorker(dbURL string, repository repository.Querier, kafka kafka.KafkaProducer) *OutboxWorker {
	return &OutboxWorker{
		dbURL:      dbURL,
		repository: repository,
		kafka:      kafka,
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
	conn, err := pgx.Connect(ctx, w.dbURL)

	if err != nil {
		slog.Error("outbox worker: failed to connect", "error", err)
		return
	}

	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, "LISTEN outbox_channel"); err != nil {
		slog.Error("outbox worker: failed to listen", "error", err)
		return
	}

	for {
		_, err := conn.WaitForNotification(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}

			slog.Error("outbox worker: notification error", "error", err)
			return
		}

		w.process(ctx)
	}
}
