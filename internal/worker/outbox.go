package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

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

	jobCh := make(chan repository.Outbox, len(events))

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for event := range jobCh {
				w.publishEvent(ctx, event)
			}
		}()
	}

	for _, event := range events {
		jobCh <- event
	}
	close(jobCh)

	wg.Wait()
}

func (w *OutboxWorker) publishEvent(ctx context.Context, event repository.Outbox) {
	var publishErr error

	for attempt := 1; attempt <= 3; attempt++ {
		publishErr = w.kafka.Publish(ctx, []byte(event.ID.String()), event.Payload)

		if publishErr == nil {
			break
		}

		slog.Warn("outbox: publish failed, retrying", "id", event.ID, "attempt", attempt, "error", publishErr)
		time.Sleep(time.Duration(attempt) * time.Second)
	}

	if publishErr != nil {
		slog.Error("outbox: publish failed after 3 attempts, marking as failed", "id", event.ID)
		w.repository.MarkOutboxEventFailed(ctx, event.ID)
		return
	}

	if err := w.repository.MarkOutboxEventProcessed(ctx, event.ID); err != nil {
		slog.Error("failed to mark outbox event processed", "id", event.ID, "error", err)
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
