package kafka

import (
	"context"
	"time"

	"github.com/sony/gobreaker"
)

type CircuitBreakerProducer struct {
	producer KafkaProducer
	cb       *gobreaker.CircuitBreaker
}

func NewCircuitBreakerProducer(producer KafkaProducer) *CircuitBreakerProducer {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "kafka-producer",
		MaxRequests: 3,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
	})

	return &CircuitBreakerProducer{
		producer: producer,
		cb:       cb,
	}
}

func (c *CircuitBreakerProducer) Publish(ctx context.Context, key, value []byte) error {
	_, err := c.cb.Execute(func() (interface{}, error) {
		return nil, c.producer.Publish(ctx, key, value)
	})

	return err
}

func (c *CircuitBreakerProducer) Close() error {
	return c.producer.Close()
}
