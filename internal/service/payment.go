package service

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/Ivantime-Kai/ecommerce-api/internal/kafka"
)

type PaymentService struct {
	producer kafka.KafkaProducer
}

func NewPaymentService(producer kafka.KafkaProducer) *PaymentService {
	return &PaymentService{
		producer: producer,
	}
}

func (s *PaymentService) HandleOrderCreated(ctx context.Context, event kafka.OrderCreatedEvent) error {
	slog.Info("payment: processing order", "order_id", event.OrderID, "amount", event.TotalAmount)

	// Simulate: luôn thành công
	completedEvent := kafka.PaymentCompletedEvent{
		OrderID:     event.OrderID,
		BuyerEmail:  event.BuyerEmail,
		TotalAmount: event.TotalAmount,
	}

	payload, _ := json.Marshal(completedEvent)

	if err := s.producer.Publish(ctx, []byte(event.OrderID), payload); err != nil {
		return err
	}

	slog.Info("payment: completed", "order_id", event.OrderID)
	return nil
}
