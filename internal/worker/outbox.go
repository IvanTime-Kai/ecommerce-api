package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/Ivantime-Kai/ecommerce-api/internal/cache"
	"github.com/Ivantime-Kai/ecommerce-api/internal/kafka"
	"github.com/Ivantime-Kai/ecommerce-api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type OutboxWorker struct {
	dbURL      string
	repository repository.Querier
	producers  map[string]kafka.KafkaProducer
	// lockCache  *cache.LockCache

	// RedLock
	redlock    *cache.RedLockClient
	instanceID string
	wg         sync.WaitGroup
}

func NewOutboxWorker(dbURL string, repository repository.Querier, producers map[string]kafka.KafkaProducer, redlock *cache.RedLockClient) *OutboxWorker {
	return &OutboxWorker{
		dbURL:      dbURL,
		repository: repository,
		producers:  producers,
		redlock:    redlock,
		instanceID: uuid.NewString(),
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
	producer, ok := w.producers[event.EventType]
	if !ok {
		slog.Error("outbox: unknown event type, marking as failed", "event_type", event.EventType, "id", event.ID)
		w.repository.MarkOutboxEventFailed(ctx, event.ID)
		return
	}

	var publishErr error

	for attempt := 1; attempt <= 3; attempt++ {
		publishErr = producer.Publish(ctx, []byte(event.ID.String()), event.Payload)

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

func (w *OutboxWorker) runOnce(ctx context.Context) error {
	conn, err := pgx.Connect(ctx, w.dbURL)

	if err != nil {
		return err
	}

	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, "LISTEN outbox_channel"); err != nil {
		return err
	}

	if mutex, ok := w.redlock.AcquireLock(ctx, "lock:outbox", 30*time.Second); ok {
		w.process(ctx)
		w.redlock.ReleaseLock(ctx, mutex)
	}

	for {
		_, err := conn.WaitForNotification(ctx)
		if err != nil {
			return nil
		}

		if mutex, ok := w.redlock.AcquireLock(ctx, "lock:outbox", 30*time.Second); ok {
			w.process(ctx)
			w.redlock.ReleaseLock(ctx, mutex)
		}
	}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			if err := w.runOnce(ctx); err != nil {
				if ctx.Err() != nil {
					return
				}

				slog.Error("outbox worker: restarting", "error", err)
				time.Sleep(5 * time.Second)
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()
}

func (w *OutboxWorker) Wait() {
	w.wg.Wait()
}
