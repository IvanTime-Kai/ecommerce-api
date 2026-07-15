package app

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/Ivantime-Kai/ecommerce-api/internal/cache"
	"github.com/Ivantime-Kai/ecommerce-api/internal/config"
	"github.com/Ivantime-Kai/ecommerce-api/internal/kafka"
	"github.com/Ivantime-Kai/ecommerce-api/internal/service"
)

type Producers struct {
	Order     kafka.KafkaProducer
	Payment   kafka.KafkaProducer
	Refund    kafka.KafkaProducer
	Inventory kafka.KafkaProducer
	CB        *kafka.CircuitBreakerProducer
}

func setupProducers(cfg *config.Config) (*Producers, error) {
	order, err := kafka.NewProducer(cfg.Kafka.Broker, kafka.TopicOrderCreated)
	if err != nil {
		return nil, err
	}
	payment, err := kafka.NewProducer(cfg.Kafka.Broker, kafka.TopicPaymentCompleted)
	if err != nil {
		return nil, err
	}
	refund, err := kafka.NewProducer(cfg.Kafka.Broker, kafka.TopicPaymentRefund)
	if err != nil {
		return nil, err
	}
	inventory, err := kafka.NewProducer(cfg.Kafka.Broker, kafka.TopicInventoryReserved)
	if err != nil {
		return nil, err
	}
	return &Producers{
		Order:     order,
		Payment:   payment,
		Refund:    refund,
		Inventory: inventory,
		CB:        kafka.NewCircuitBreakerProducer(order),
	}, nil
}

func (p *Producers) Close() {
	p.Order.Close()
	p.Payment.Close()
	p.Refund.Close()
	p.Inventory.Close()
}

type Consumers struct {
	payment                *kafka.ConsumerGroup
	inventoryFailedPayment *kafka.ConsumerGroup
	inventory              *kafka.ConsumerGroup
	inventoryEvent         *kafka.ConsumerGroup
	inventoryFailed        *kafka.ConsumerGroup
	orderEvent             *kafka.ConsumerGroup
	paymentFailed          *kafka.ConsumerGroup
}

func setupConsumers(
	ctx context.Context,
	cfg *config.Config,
	paymentSvc *service.PaymentService,
	orderSvc *service.OrderService,
	inventorySvc *service.InventoryService,
	idempotencyCache *cache.IdempotencyCache,
	notificationSvc *service.NotificationService,
) *Consumers {
	c := &Consumers{}

	c.payment = kafka.NewConsumerGroup(cfg.Kafka.Broker, kafka.TopicOrderCreated, "payment-service", 3)
	c.payment.Consume(ctx, func(key, value []byte) error {
		var event kafka.OrderCreatedEvent
		if err := json.Unmarshal(value, &event); err != nil {
			return err
		}
		return paymentSvc.HandleOrderCreated(ctx, event)
	})

	c.inventoryFailedPayment = kafka.NewConsumerGroup(cfg.Kafka.Broker, kafka.TopicInventoryFailed, "payment-service-refund", 3)
	c.inventoryFailedPayment.Consume(ctx, func(key, value []byte) error {
		var event kafka.InventoryFailedEvent
		if err := json.Unmarshal(value, &event); err != nil {
			return err
		}
		return paymentSvc.HandleInventoryFailed(ctx, event)
	})

	c.inventory = kafka.NewConsumerGroup(cfg.Kafka.Broker, kafka.TopicPaymentCompleted, "inventory-service", 3)
	c.inventory.Consume(ctx, func(key, value []byte) error {
		var event kafka.PaymentCompletedEvent
		if err := json.Unmarshal(value, &event); err != nil {
			return err
		}
		return inventorySvc.HandlePaymentCompleted(ctx, event)
	})

	c.inventoryEvent = kafka.NewConsumerGroup(cfg.Kafka.Broker, kafka.TopicInventoryReserved, "order-service-inventory", 3)
	c.inventoryEvent.Consume(ctx, func(key, value []byte) error {
		var event kafka.InventoryReservedEvent
		if err := json.Unmarshal(value, &event); err != nil {
			return err
		}
		return orderSvc.HandleInventoryReserved(ctx, event)
	})

	c.inventoryFailed = kafka.NewConsumerGroup(cfg.Kafka.Broker, kafka.TopicInventoryFailed, "order-service-inventory-failed", 3)
	c.inventoryFailed.Consume(ctx, func(key, value []byte) error {
		var event kafka.InventoryFailedEvent
		if err := json.Unmarshal(value, &event); err != nil {
			return err
		}
		return orderSvc.HandleInventoryFailed(ctx, event)
	})

	c.orderEvent = kafka.NewConsumerGroup(cfg.Kafka.Broker, kafka.TopicPaymentCompleted, "order-service-payment", 3)
	c.orderEvent.Consume(ctx, func(key, value []byte) error {
		var event kafka.PaymentCompletedEvent
		if err := json.Unmarshal(value, &event); err != nil {
			return err
		}
		processed, err := idempotencyCache.IsProcessed(ctx, kafka.TopicPaymentCompleted, event.OrderID)
		if err != nil {
			return err
		}
		if processed {
			slog.Info("skipping duplicate event", "order_id", event.OrderID)
			return nil
		}
		if err := orderSvc.HandlePaymentCompleted(ctx, event); err != nil {
			return err
		}
		if err := idempotencyCache.MarkProcessed(ctx, kafka.TopicPaymentCompleted, event.OrderID); err != nil {
			return err
		}
		if err := notificationSvc.SendOrderConfirmation(event); err != nil {
			slog.Warn("failed to send order confirmation email", "order_id", event.OrderID, "error", err)
		}
		return nil
	})

	c.paymentFailed = kafka.NewConsumerGroup(cfg.Kafka.Broker, kafka.TopicPaymentFailed, "order-service-payment-failed", 3)
	c.paymentFailed.Consume(ctx, func(key, value []byte) error {
		var event kafka.PaymentFailedEvent
		if err := json.Unmarshal(value, &event); err != nil {
			return err
		}
		return orderSvc.HandlePaymentFailed(ctx, event)
	})

	return c
}

func (c *Consumers) Close() {
	c.payment.Close()
	c.inventoryFailedPayment.Close()
	c.inventory.Close()
	c.inventoryEvent.Close()
	c.inventoryFailed.Close()
	c.orderEvent.Close()
	c.paymentFailed.Close()
}

func (c *Consumers) Wait() {
	c.payment.Wait()
	c.inventoryFailedPayment.Wait()
	c.inventory.Wait()
	c.inventoryEvent.Wait()
	c.inventoryFailed.Wait()
	c.orderEvent.Wait()
	c.paymentFailed.Wait()
}
