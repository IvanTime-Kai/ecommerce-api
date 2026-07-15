package kafka

import (
	"context"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader      *kafka.Reader
	dlqProducer KafkaProducer
}

func NewConsumer(brokerAddress, topic, groupID string, dlqProducer KafkaProducer) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokerAddress},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
		MaxWait:  100 * time.Millisecond,
	})

	return &Consumer{
		reader:      reader,
		dlqProducer: dlqProducer,
	}
}

func (c *Consumer) Consume(ctx context.Context, handler func(key, value []byte) error) error {
	for {

		message, err := c.reader.ReadMessage(ctx)

		if err != nil {
			return err
		}

		if err := c.handleWithRetry(ctx, message, handler); err != nil {
			slog.Error("message failed after retries, sending to DLQ", "error", err)
			c.sendToDLQ(ctx, message)
		}

	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func (c *Consumer) handleWithRetry(ctx context.Context, msg kafka.Message, handler func(key, value []byte) error) error {
	backoff := 100 * time.Millisecond
	var err error
	for attempt := 1; attempt <= 3; attempt++ {
		err = handler(msg.Key, msg.Value)
		if err == nil {
			return nil
		}
		slog.Warn("handler failed, retrying", "attempt", attempt, "error", err)
		if attempt < 3 {
			time.Sleep(backoff)
			backoff *= 2
		}
	}
	return err
}
func (c *Consumer) sendToDLQ(ctx context.Context, msg kafka.Message) {
	if c.dlqProducer == nil {
		return
	}

	msg.Headers = append(msg.Headers, kafka.Header{
		Key:   "source-topic",
		Value: []byte(msg.Topic),
	})

	if err := c.dlqProducer.Publish(ctx, msg.Key, msg.Value); err != nil {
		slog.Error("failed to send message to DLQ", "error", err)
	}
}
