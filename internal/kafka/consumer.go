package kafka

import (
	"context"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokerAddress, topic, groupID string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokerAddress},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
		MaxWait:  100 * time.Millisecond,
	})

	return &Consumer{
		reader: reader,
	}
}

func (c *Consumer) Consume(ctx context.Context, handler func(key, value []byte) error) error {
	for {

		message, err := c.reader.ReadMessage(ctx)

		if err != nil {
			return err
		}

		err = handler(message.Key, message.Value)

		if err != nil {
			slog.Error("failed to handle message", "error", err)
			continue
		}

	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
