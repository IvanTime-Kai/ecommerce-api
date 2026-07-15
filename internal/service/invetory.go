package service

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/Ivantime-Kai/ecommerce-api/internal/cache"
	"github.com/Ivantime-Kai/ecommerce-api/internal/kafka"
	"github.com/Ivantime-Kai/ecommerce-api/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InventoryService struct {
	repository repository.Querier
	db         *pgxpool.Pool
	producer   kafka.KafkaProducer
	cache      *cache.StockCache
}

func NewInventoryService(repo repository.Querier, db *pgxpool.Pool, producer kafka.KafkaProducer, cache *cache.StockCache) *InventoryService {
	return &InventoryService{
		repository: repo,
		db:         db,
		producer:   producer,
		cache:      cache,
	}
}

func (s *InventoryService) HandlePaymentCompleted(ctx context.Context, event kafka.PaymentCompletedEvent) error {
	slog.Info("inventory: reserving stock", "order_id", event.OrderID)

	orderID, err := uuid.Parse(event.OrderID)
	if err != nil {
		return err
	}

	items, err := s.repository.GetOrderItemsByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	deducted := []repository.OrderItem{}

	// Thử trừ stock từng item
	for _, item := range items {
		_, err := s.repository.DeductProductStock(ctx, repository.DeductProductStockParams{
			ID:    item.ProductID,
			Stock: item.Quantity,
		})

		if err != nil {

			for _, d := range deducted {
				s.repository.RestoreProductStock(ctx, repository.RestoreProductStockParams{
					ID:    d.ProductID,
					Stock: d.Quantity,
				})
				s.cache.RestoreStock(ctx, d.ProductID.String(), int(d.Quantity))
			}

			payload, _ := json.Marshal(kafka.InventoryFailedEvent{
				OrderID: event.OrderID,
				Reason:  err.Error(),
			})
			s.producer.Publish(ctx, []byte(event.OrderID), payload)
			return nil
		}

		s.cache.DeductStock(ctx, item.ProductID.String(), int(item.Quantity))
		deducted = append(deducted, item)
	}

	payload, _ := json.Marshal(kafka.InventoryReservedEvent{
		OrderID: event.OrderID,
	})

	s.producer.Publish(ctx, []byte(event.OrderID), payload)
	slog.Info("inventory: stock reserved", "order_id", event.OrderID)
	return nil
}
